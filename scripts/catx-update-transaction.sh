#!/usr/bin/env bash

# CatX-UI transactional update module. This file is published as a checksummed
# release asset and sourced only by update.sh after verification.

readonly CATX_RELEASE_OWNER="CatCodeArbelin"
readonly CATX_RELEASE_REPOSITORY="CatX-UI"
readonly CATX_RELEASE_SLUG="${CATX_RELEASE_OWNER}/${CATX_RELEASE_REPOSITORY}"
readonly CATX_ASSET_PREFIX="catx-ui"
readonly CATX_DEV_RELEASE_TAG="dev-latest"
readonly CATX_RELEASE_WEB="https://github.com/${CATX_RELEASE_SLUG}"

catx_transaction_dir=""
catx_transaction_active=0
catx_transaction_tag=""
catx_rollback_healthy=0
catx_rollback_attempted=0
catx_service_quiesced=0

_catx_log() {
    printf '%s\n' "$*"
}

_catx_audit() {
    local outcome="$1" detail="$2"
    local audit_file="${CATX_UPDATE_AUDIT_FILE:-/var/log/x-ui/update-audit.log}"
    local audit_dir
    audit_dir=$(dirname "$audit_file")
    mkdir -p "$audit_dir" 2> /dev/null || return 0
    touch "$audit_file" 2> /dev/null || return 0
    chmod 600 "$audit_file" 2> /dev/null || true
    printf '%s run_id=%s tag=%s outcome=%s rollback_healthy=%s detail=%q\n' \
        "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "${xui_update_run_id:-0}" \
        "${catx_transaction_tag:-unknown}" "$outcome" "$catx_rollback_healthy" "$detail" >> "$audit_file"
}

_catx_validate_tag() {
    [[ "$1" == "${CATX_DEV_RELEASE_TAG}" || "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

_catx_resolve_tag() {
    if [[ -n "${XUI_UPDATE_TAG:-}" ]]; then
        _catx_validate_tag "${XUI_UPDATE_TAG}" || return 1
        printf '%s\n' "${XUI_UPDATE_TAG}"
        return 0
    fi
    local url tag
    url=$(${curl_bin:-curl} -sSLI -o /dev/null -w '%{url_effective}' --retry 5 --retry-delay 3 \
        --connect-timeout 15 --max-time 60 "${CATX_RELEASE_WEB}/releases/latest" 2> /dev/null || true)
    tag=${url##*/tag/}
    if [[ "$tag" != "$url" && "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        printf '%s\n' "$tag"
        return 0
    fi
    tag=$(${curl_bin:-curl} -fsSL --retry 5 --retry-delay 3 --connect-timeout 15 --max-time 60 \
        "https://api.github.com/repos/${CATX_RELEASE_SLUG}/releases/latest" 2> /dev/null |
        grep '"tag_name":' | head -n1 | sed -E 's/.*"([^"]+)".*/\1/' || true)
    [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || return 1
    printf '%s\n' "$tag"
}

_catx_download_verified() {
    local tag="$1" name="$2" dest="$3"
    local url="${CATX_RELEASE_WEB}/releases/download/${tag}/${name}"
    local sums="${dest}.sha256" expected actual recorded_name code
    rm -f -- "$dest" "$sums"
    ${curl_bin:-curl} -fLR --retry 5 --retry-delay 3 --connect-timeout 15 --speed-limit 1 --speed-time 300 \
        -o "$dest" "$url" || return 1
    code=$(${curl_bin:-curl} -sL --retry 5 --retry-delay 3 --connect-timeout 15 --max-time 60 \
        -o "$sums" -w '%{http_code}' "${url}.sha256" 2> /dev/null)
    if [[ "$code" != "200" ]]; then
        rm -f -- "$dest" "$sums"
        _catx_log "CatX-UI checksum download failed for ${name} (HTTP ${code})"
        return 1
    fi
    [[ $(grep -cve '^[[:space:]]*$' "$sums") -eq 1 ]] || {
        rm -f -- "$dest" "$sums"
        return 1
    }
    expected=$(awk 'NF {print $1}' "$sums")
    recorded_name=$(awk 'NF {print $2}' "$sums")
    actual=$(sha256sum "$dest" | awk '{print $1}')
    rm -f -- "$sums"
    if [[ ! "$expected" =~ ^[0-9a-f]{64}$ || "$recorded_name" != "$name" || "$expected" != "$actual" ]]; then
        rm -f -- "$dest"
        _catx_log "CatX-UI integrity verification failed for ${name}"
        return 1
    fi
    _catx_log "Checksum verified for ${name}: ${actual}"
}

_catx_safe_directory() {
    local target="$1"
    [[ "$target" == /* ]] || return 1
    case "$target" in
        / | /bin | /boot | /dev | /etc | /home | /lib | /lib64 | /opt | /proc | /root | /run | /sbin | /srv | /sys | /tmp | /usr | /var)
            return 1
            ;;
    esac
    return 0
}

_catx_validate_archive_paths() {
    local archive="$1" entry
    while IFS= read -r entry; do
        [[ -n "$entry" ]] || continue
        [[ "$entry" == x-ui || "$entry" == x-ui/* ]] || return 1
        [[ "$entry" != /* && "$entry" != *"../"* && "$entry" != *"/.." ]] || return 1
    done < <(tar -tzf "$archive") || return 1
    # Only regular files and directories are needed by the release. Reject
    # links and device/special entries before extraction so they cannot write
    # through a crafted target while the archive is being unpacked as root.
    while IFS= read -r entry; do
        case "${entry:0:1}" in
            - | d) ;;
            *) return 1 ;;
        esac
    done < <(tar -tvzf "$archive") || return 1
}

_catx_validate_candidate() {
    local candidate="$1" tag="$2" identity
    [[ -x "$candidate/x-ui" && -s "$candidate/x-ui.sh" ]] || return 1
    if find "$candidate" -type l -print -quit | grep -q .; then
        _catx_log "Candidate contains symbolic links; refusing release payload"
        return 1
    fi
    chmod +x "$candidate/x-ui" "$candidate/x-ui.sh"
    identity=$("$candidate/x-ui" release-info 2> /dev/null) || return 1
    grep -Fxq "product=CatX-UI" <<< "$identity" || return 1
    grep -Fxq "repository=${CATX_RELEASE_SLUG}" <<< "$identity" || return 1
    if [[ "$tag" == "${CATX_DEV_RELEASE_TAG}" ]]; then
        grep -Fxq "channel=dev" <<< "$identity" || return 1
    else
        grep -Fxq "channel=stable" <<< "$identity" || return 1
        grep -Fxq "fork_version=${tag#v}" <<< "$identity" || return 1
    fi
}

_catx_stop_service() {
    local stop_rc=0
    if [[ "${release:-}" == "alpine" ]]; then
        rc-service x-ui stop > /dev/null 2>&1 || stop_rc=$?
    else
        systemctl stop x-ui > /dev/null 2>&1 || stop_rc=$?
    fi
    pkill -f 'mtg-linux-[^ ]* run ' > /dev/null 2>&1 || true
    pkill -f 'tuic-server.*-c .*bin/tuic/tuic_[0-9]+\.json' > /dev/null 2>&1 || true
    return "$stop_rc"
}

_catx_start_service() {
    if [[ "${release:-}" == "alpine" ]]; then
        rc-update add x-ui > /dev/null 2>&1 || return 1
        rc-service x-ui start > /dev/null 2>&1
    else
        systemctl daemon-reload > /dev/null 2>&1 || return 1
        systemctl enable x-ui > /dev/null 2>&1 || return 1
        systemctl start x-ui > /dev/null 2>&1
    fi
}

_catx_service_healthy() {
    local attempts="${CATX_HEALTHCHECK_ATTEMPTS:-15}" healthy=0
    while ((attempts > 0)); do
        if [[ "${release:-}" == "alpine" ]]; then
            rc-service x-ui status > /dev/null 2>&1 && healthy=$((healthy + 1)) || healthy=0
        else
            systemctl is-active --quiet x-ui && healthy=$((healthy + 1)) || healthy=0
        fi
        ((healthy >= 2)) && return 0
        attempts=$((attempts - 1))
        sleep 1
    done
    return 1
}

_catx_resume_known_good() {
    if _catx_start_service && _catx_service_healthy; then
        catx_service_quiesced=0
        _catx_audit aborted "update stopped before activation; previous installation remains healthy"
        return 0
    fi
    _catx_audit rollback-failed "could not resume previous installation before activation"
    _catx_log "CRITICAL: previous installation could not be resumed after the update aborted"
    return 1
}

_catx_db_kind() {
    case "${XUI_DB_TYPE:-sqlite}" in
        postgres | postgresql | pg) echo postgres ;;
        *) echo sqlite ;;
    esac
}

_catx_snapshot_database() {
    local backup="$1" kind
    kind=$(_catx_db_kind)
    printf '%s\n' "$kind" > "$backup/db-kind"
    if [[ "$kind" == "postgres" ]]; then
        [[ -n "${XUI_DB_DSN:-}" ]] || return 1
        command -v pg_dump > /dev/null 2>&1 && command -v pg_restore > /dev/null 2>&1 || return 1
        pg_dump --format=custom --file="$backup/database.dump" "${XUI_DB_DSN}" > /dev/null
        return
    fi
    local db_folder="${XUI_DB_FOLDER:-/etc/x-ui}"
    _catx_safe_directory "$db_folder" || return 1
    mkdir -p "$backup/database"
    [[ ! -e "$db_folder" ]] || cp -a "$db_folder/." "$backup/database/"
}

_catx_restore_database() {
    local backup="$1" kind db_folder failed_state restore_state
    kind=$(<"$backup/db-kind")
    if [[ "$kind" == "postgres" ]]; then
        pg_restore --clean --if-exists --single-transaction --no-owner --dbname="${XUI_DB_DSN}" "$backup/database.dump" > /dev/null
        return
    fi
    db_folder="${XUI_DB_FOLDER:-/etc/x-ui}"
    _catx_safe_directory "$db_folder" || return 1
    failed_state="${catx_transaction_dir}/failed-database"
    # A sibling of the live directory is on the same filesystem, allowing the
    # final rename to be atomic. The parent may legitimately be /etc; only the
    # actual database directory is ever a replace/delete target.
    restore_state="${db_folder}.catx-restore.$$"
    rm -rf -- "$failed_state"
    rm -rf -- "$restore_state"
    mkdir -p "$restore_state" || return 1
    cp -a "$backup/database/." "$restore_state/" || {
        rm -rf -- "$restore_state"
        return 1
    }
    [[ ! -e "$db_folder" ]] || mv "$db_folder" "$failed_state" || return 1
    mv "$restore_state" "$db_folder" || {
        [[ ! -e "$db_folder" && -e "$failed_state" ]] && mv "$failed_state" "$db_folder"
        return 1
    }
    rm -rf -- "$failed_state"
}

_catx_external_paths() {
    printf '%s\n' "${CATX_CLI_PATH:-/usr/bin/x-ui}"
    if [[ "${release:-}" == "alpine" ]]; then
        printf '%s\n' "/etc/init.d/x-ui"
    else
        printf '%s\n' "${xui_service:-/etc/systemd/system}/x-ui.service"
    fi
    if [[ -n "${CATX_ENV_FILE_PATHS:-}" ]]; then
        tr ':' '\n' <<< "${CATX_ENV_FILE_PATHS}"
    else
        printf '%s\n' "/etc/default/x-ui" "/etc/conf.d/x-ui" "/etc/sysconfig/x-ui"
    fi
}

_catx_snapshot_external() {
    local backup="$1" item encoded
    mkdir -p "$backup/external"
    : > "$backup/external-manifest"
    while IFS= read -r item; do
        encoded=${item#/}
        if [[ -e "$item" ]]; then
            mkdir -p "$backup/external/$(dirname "$encoded")"
            cp -a "$item" "$backup/external/$encoded" || return 1
            printf 'present %s\n' "$item" >> "$backup/external-manifest"
        else
            printf 'absent %s\n' "$item" >> "$backup/external-manifest"
        fi
    done < <(_catx_external_paths)
}

_catx_restore_external() {
    local backup="$1" state item encoded
    while read -r state item; do
        encoded=${item#/}
        rm -f -- "$item"
        if [[ "$state" == "present" ]]; then
            mkdir -p "$(dirname "$item")"
            cp -a "$backup/external/$encoded" "$item" || return 1
        fi
    done < "$backup/external-manifest"
}

_catx_install_service_unit() {
    local live="$1" source dest temp
    if [[ "${release:-}" == "alpine" ]]; then
        source="$live/x-ui.rc"
        dest="/etc/init.d/x-ui"
    else
        dest="${xui_service:-/etc/systemd/system}/x-ui.service"
        case "${release:-}" in
            ubuntu | debian | armbian) source="$live/x-ui.service.debian" ;;
            arch | manjaro | parch) source="$live/x-ui.service.arch" ;;
            *) source="$live/x-ui.service.rhel" ;;
        esac
        [[ -s "$live/x-ui.service" ]] && source="$live/x-ui.service"
    fi
    [[ -s "$source" ]] || return 1
    mkdir -p "$(dirname "$dest")"
    temp="${dest}.catx-new.$$"
    cp -f "$source" "$temp" || return 1
    chmod 755 "$temp"
    [[ "${release:-}" == "alpine" ]] || chmod 644 "$temp"
    mv -f "$temp" "$dest"
}

_catx_restore_custom_bin() {
    local old_bin="$1" new_bin="$2" file rel
    [[ -d "$old_bin" ]] || return 0
    while IFS= read -r -d '' file; do
        rel=${file#"$old_bin"/}
        case "$rel" in
            config.json | mtproto | mtproto/* | tuic | tuic/*) continue ;;
        esac
        if [[ ! -e "$new_bin/$rel" ]]; then
            mkdir -p "$new_bin/$(dirname "$rel")"
            cp -a "$file" "$new_bin/$rel" || return 1
        fi
    done < <(find "$old_bin" \( -type f -o -type l \) -print0)
}

_catx_prepare_arch_binaries() {
    local live="$1" platform
    platform=$(arch)
    if [[ "$platform" == "armv5" || "$platform" == "armv6" || "$platform" == "armv7" ]]; then
        [[ ! -e "$live/bin/xray-linux-$platform" ]] || mv "$live/bin/xray-linux-$platform" "$live/bin/xray-linux-arm32"
        [[ ! -e "$live/bin/mtg-linux-$platform" ]] || mv "$live/bin/mtg-linux-$platform" "$live/bin/mtg-linux-arm"
    fi
    chmod +x "$live/x-ui"
    find "$live/bin" -maxdepth 1 -type f \( -name 'xray-linux-*' -o -name 'mtg-linux-*' -o -name 'tuic-server' \) -exec chmod +x {} +
}

_catx_rollback() {
    local backup="$catx_transaction_dir/backup" live="${xui_folder:-/usr/local/x-ui}"
    catx_rollback_healthy=0
    catx_rollback_attempted=1
    _catx_log "Update failed; restoring the previous known-good installation"
    _catx_stop_service || return 1
    if [[ -d "$backup/live" ]]; then
        if [[ -e "$live" ]]; then
            _catx_safe_directory "$live" || return 1
            rm -rf -- "$catx_transaction_dir/failed-candidate"
            mv "$live" "$catx_transaction_dir/failed-candidate" || return 1
        fi
        mv "$backup/live" "$live" || return 1
    elif [[ ! -x "$live/x-ui" ]]; then
        # A prior rollback attempt consumed backup/live but did not leave a
        # runnable installation. Do not remove anything else on a retry.
        return 1
    fi
    _catx_restore_external "$backup" || return 1
    _catx_restore_database "$backup" || return 1
    _catx_start_service || return 1
    _catx_service_healthy || return 1
    catx_rollback_healthy=1
    catx_transaction_active=0
    catx_service_quiesced=0
    _catx_audit rollback "previous installation, configuration, and database restored"
    return 0
}

_catx_fail_and_rollback() {
    local reason="$1"
    _catx_audit failure "$reason"
    if ! _catx_rollback; then
        _catx_audit rollback-failed "$reason"
        _catx_log "CRITICAL: automatic rollback did not return the previous installation to a healthy state"
        return 3
    fi
    _catx_log "Rollback completed and the previous installation is healthy"
    return 2
}

catx_apply_staged_update() {
    local candidate="$1" tag="$2" live="${xui_folder:-/usr/local/x-ui}"
    local backup="$catx_transaction_dir/backup" cli="${CATX_CLI_PATH:-/usr/bin/x-ui}"
    _catx_safe_directory "$live" || return 1
    [[ -d "$live" && -x "$live/x-ui" ]] || return 1

    mkdir -p "$backup"
    # Mark the quiesced window before stopping so TERM/INT during the stop can
    # still restart and healthcheck the known-good installation.
    catx_service_quiesced=1
    if ! _catx_stop_service; then
        catx_service_quiesced=0
        return 1
    fi
    _catx_snapshot_database "$backup" || {
        _catx_resume_known_good || return 3
        return 1
    }
    _catx_snapshot_external "$backup" || {
        _catx_resume_known_good || return 3
        return 1
    }
    mv "$live" "$backup/live" || {
        _catx_resume_known_good || return 3
        return 1
    }
    catx_transaction_active=1

    _catx_restore_custom_bin "$backup/live/bin" "$candidate/bin" || {
        _catx_fail_and_rollback "restore custom bin files into candidate"
        return $?
    }
    mv "$candidate" "$live" || {
        _catx_fail_and_rollback "activate staged installation"
        return $?
    }
    _catx_prepare_arch_binaries "$live" || {
        _catx_fail_and_rollback "prepare architecture-specific binaries"
        return $?
    }
    mkdir -p "$(dirname "$cli")"
    cp -f "$live/x-ui.sh" "${cli}.catx-new.$$" && chmod 755 "${cli}.catx-new.$$" && mv -f "${cli}.catx-new.$$" "$cli" || {
        rm -f "${cli}.catx-new.$$"
        _catx_fail_and_rollback "install command-line script"
        return $?
    }
    _catx_install_service_unit "$live" || {
        _catx_fail_and_rollback "install service unit"
        return $?
    }
    if ! "$live/x-ui" migrate; then
        _catx_fail_and_rollback "database migration failed"
        return $?
    fi
    _catx_start_service || {
        _catx_fail_and_rollback "candidate service failed to start"
        return $?
    }
    _catx_service_healthy || {
        _catx_fail_and_rollback "candidate healthcheck failed"
        return $?
    }

    catx_transaction_active=0
    catx_service_quiesced=0
    _catx_audit commit "candidate installed, migrated, and healthy"
    rm -rf -- "$catx_transaction_dir"
    catx_transaction_dir=""
    _catx_log "CatX-UI ${tag} update committed"
    return 0
}

catx_transactional_update() {
    local tag parent archive archive_name candidate
    tag=$(_catx_resolve_tag) || {
        _catx_log "Could not resolve a trusted CatX-UI release"
        return 1
    }
    _catx_validate_tag "$tag" || return 1
    catx_transaction_tag="$tag"
    parent=$(dirname "${xui_folder:-/usr/local/x-ui}")
    _catx_safe_directory "$parent" || return 1
    catx_transaction_dir=$(mktemp -d "$parent/.catx-update.XXXXXX") || return 1
    chmod 700 "$catx_transaction_dir"
    archive_name="${CATX_ASSET_PREFIX}-linux-$(arch).tar.gz"
    archive="$catx_transaction_dir/$archive_name"

    _catx_download_verified "$tag" "$archive_name" "$archive" || {
        rm -rf -- "$catx_transaction_dir"
        return 1
    }
    _catx_validate_archive_paths "$archive" || {
        _catx_log "Release archive contains an unsafe path"
        rm -rf -- "$catx_transaction_dir"
        return 1
    }
    mkdir -p "$catx_transaction_dir/stage"
    tar -xzf "$archive" --no-same-owner --no-same-permissions -C "$catx_transaction_dir/stage" || {
        rm -rf -- "$catx_transaction_dir"
        return 1
    }
    candidate="$catx_transaction_dir/stage/x-ui"
    _catx_validate_candidate "$candidate" "$tag" || {
        _catx_log "Release archive is not a valid ${CATX_RELEASE_SLUG} candidate for ${tag}"
        rm -rf -- "$catx_transaction_dir"
        return 1
    }

    catx_apply_staged_update "$candidate" "$tag"
}

catx_update_exit_guard() {
    local code="$1"
    if [[ "$code" -ne 0 && "$catx_transaction_active" -eq 1 ]]; then
        _catx_fail_and_rollback "update interrupted with exit code ${code}" || true
    elif [[ "$code" -ne 0 && "$catx_service_quiesced" -eq 1 ]]; then
        # If the atomic live->backup rename completed just before a signal was
        # delivered, promote this to the full rollback path. Otherwise no live
        # files changed and only the old service needs to be resumed.
        if [[ -d "${catx_transaction_dir}/backup/live" ]]; then
            catx_transaction_active=1
            _catx_fail_and_rollback "update interrupted at activation boundary with exit code ${code}" || true
        else
            _catx_resume_known_good || true
        fi
    fi
}
