// Package forkrelease owns CatX-UI's distribution identity and release trust
// policy. It intentionally does not depend on panel services so shell and web
// updater integration points can remain thin.
package forkrelease

import (
	_ "embed"
	"fmt"
	"strings"
)

// identity.env is the authoritative distribution identity. Shell entry points
// mirror these values because they must also work as standalone downloaded
// files; release verification tests reject any drift from this source.
//
//go:embed identity.env
var identityEnv string

//go:embed fork_version
var forkVersion string

//go:embed upstream_version
var upstreamVersion string

type Identity struct {
	ProductName        string
	ReleaseOwner       string
	ReleaseRepository  string
	AssetPrefix        string
	DevReleaseTag      string
	BundledXrayVersion string
}

var Current = mustParseIdentity(identityEnv)

func mustParseIdentity(data string) Identity {
	values := make(map[string]string)
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			panic(fmt.Sprintf("invalid fork release identity line %q", line))
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	required := func(key string) string {
		value := values[key]
		if value == "" {
			panic("missing fork release identity value " + key)
		}
		return value
	}
	return Identity{
		ProductName:        required("CATX_PRODUCT_NAME"),
		ReleaseOwner:       required("CATX_RELEASE_OWNER"),
		ReleaseRepository:  required("CATX_RELEASE_REPOSITORY"),
		AssetPrefix:        required("CATX_ASSET_PREFIX"),
		DevReleaseTag:      required("CATX_DEV_RELEASE_TAG"),
		BundledXrayVersion: required("CATX_XRAY_VERSION"),
	}
}

func ForkVersion() string {
	return strings.TrimSpace(forkVersion)
}

func UpstreamBaseVersion() string {
	return strings.TrimSpace(upstreamVersion)
}

func (i Identity) RepositorySlug() string {
	return i.ReleaseOwner + "/" + i.ReleaseRepository
}

func (i Identity) APIRepositoryURL() string {
	return "https://api.github.com/repos/" + i.RepositorySlug()
}

func (i Identity) WebRepositoryURL() string {
	return "https://github.com/" + i.RepositorySlug()
}

func (i Identity) RawRepositoryURL() string {
	return "https://raw.githubusercontent.com/" + i.RepositorySlug()
}

func (i Identity) LinuxArchiveName(arch string) string {
	return i.AssetPrefix + "-linux-" + arch + ".tar.gz"
}

func (i Identity) UpdateScriptAssetName() string {
	return i.AssetPrefix + "-update.sh"
}

func (i Identity) InstallScriptAssetName() string {
	return i.AssetPrefix + "-install.sh"
}

func (i Identity) MenuScriptAssetName() string {
	return i.AssetPrefix + ".sh"
}
