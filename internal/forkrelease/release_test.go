package forkrelease

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderFetchesCatXStableReleaseFromFakeServer(t *testing.T) {
	payload := []byte("#!/bin/bash\necho catx\n")
	digest := sha256.Sum256(payload)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := AssetBaseForTest(server.URL, Current)
		switch r.URL.Path {
		case "/api/releases/latest":
			fmt.Fprintf(w, `{"url":%q,"html_url":%q,"tag_name":"v0.2.0","draft":false,"prerelease":false,"assets":[{"name":"catx-ui-update.sh","browser_download_url":%q,"size":%d},{"name":"catx-ui-update.sh.sha256","browser_download_url":%q,"size":84}]}`,
				server.URL+"/api/releases/1", base+"/releases/tag/v0.2.0", base+"/releases/download/v0.2.0/catx-ui-update.sh", len(payload), base+"/releases/download/v0.2.0/catx-ui-update.sh.sha256")
		case "/" + Current.RepositorySlug() + "/releases/download/v0.2.0/catx-ui-update.sh":
			_, _ = w.Write(payload)
		case "/" + Current.RepositorySlug() + "/releases/download/v0.2.0/catx-ui-update.sh.sha256":
			fmt.Fprintf(w, "%x  catx-ui-update.sh\n", digest)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	base := AssetBaseForTest(server.URL, Current)
	provider := Provider{Identity: Current, Client: server.Client(), APIBaseURL: server.URL + "/api", WebBaseURL: base, AssetBaseURL: base}
	release, err := provider.Fetch(context.Background(), ChannelStable)
	if err != nil {
		t.Fatalf("fetch release: %v", err)
	}
	got, err := provider.DownloadVerified(context.Background(), release, Current.UpdateScriptAssetName(), 1<<20)
	if err != nil {
		t.Fatalf("download verified: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}
}

func TestProviderRejectsOfficialUpstreamReleaseAndAssets(t *testing.T) {
	provider := NewProvider(http.DefaultClient)
	release := &Release{
		APIURL:     "https://api.github.com/repos/MHSanaei/3x-ui/releases/1",
		HTMLURL:    "https://github.com/MHSanaei/3x-ui/releases/tag/v3.8.5",
		TagName:    "v3.8.5",
		Prerelease: false,
	}
	if err := provider.ValidateRelease(release, ChannelStable); err == nil || !strings.Contains(err.Error(), "trusted repository") {
		t.Fatalf("wrong-repository release error = %v, want trusted-repository rejection", err)
	}

	release.APIURL = Current.APIRepositoryURL() + "/releases/1"
	release.HTMLURL = Current.WebRepositoryURL() + "/releases/tag/v3.8.5"
	release.Assets = []Asset{{Name: Current.UpdateScriptAssetName(), BrowserDownloadURL: "https://github.com/MHSanaei/3x-ui/releases/download/v3.8.5/catx-ui-update.sh", Size: 10}}
	if _, err := provider.FindAsset(release, Current.UpdateScriptAssetName()); err == nil || !strings.Contains(err.Error(), "trusted repository") {
		t.Fatalf("wrong-repository asset error = %v, want trusted-repository rejection", err)
	}
}

func TestProviderEnforcesChannels(t *testing.T) {
	p := NewProvider(http.DefaultClient)
	stable := &Release{APIURL: p.APIBaseURL + "/releases/1", HTMLURL: p.WebBaseURL + "/releases/tag/v0.2.0", TagName: "v0.2.0"}
	if err := p.ValidateRelease(stable, ChannelStable); err != nil {
		t.Fatalf("stable release rejected: %v", err)
	}
	stable.Prerelease = true
	if err := p.ValidateRelease(stable, ChannelStable); err == nil {
		t.Fatal("stable channel accepted a prerelease")
	}
	stable.TagName = p.Identity.RCReleaseTag()
	stable.HTMLURL = p.WebBaseURL + "/releases/tag/" + stable.TagName
	if err := p.ValidateRelease(stable, ChannelStable); err == nil {
		t.Fatal("stable channel accepted the explicit RC tag")
	}

	dev := &Release{APIURL: p.APIBaseURL + "/releases/2", HTMLURL: p.WebBaseURL + "/releases/tag/" + Current.DevReleaseTag, TagName: Current.DevReleaseTag, Prerelease: true}
	if err := p.ValidateRelease(dev, ChannelDev); err != nil {
		t.Fatalf("dev release rejected: %v", err)
	}
	dev.TagName = "nightly"
	if err := p.ValidateRelease(dev, ChannelDev); err == nil {
		t.Fatal("dev channel accepted an unexpected rolling tag")
	}

	rc := &Release{APIURL: p.APIBaseURL + "/releases/3", HTMLURL: p.WebBaseURL + "/releases/tag/" + p.Identity.RCReleaseTag(), TagName: p.Identity.RCReleaseTag(), Prerelease: true}
	if err := p.ValidateRelease(rc, ChannelRC); err != nil {
		t.Fatalf("RC release rejected: %v", err)
	}
	rc.Prerelease = false
	if err := p.ValidateRelease(rc, ChannelRC); err == nil {
		t.Fatal("RC channel accepted a non-prerelease release")
	}
	rc.Prerelease = true
	rc.TagName = "v0.2.0-rc.3"
	rc.HTMLURL = p.WebBaseURL + "/releases/tag/" + rc.TagName
	if err := p.ValidateRelease(rc, ChannelRC); err != nil {
		t.Fatalf("approved future RC release rejected: %v", err)
	}
	rc.TagName = "v0.2.0-rc"
	rc.HTMLURL = p.WebBaseURL + "/releases/tag/" + rc.TagName
	if err := p.ValidateRelease(rc, ChannelRC); err == nil {
		t.Fatal("RC channel accepted a malformed RC tag")
	}
}

func TestProviderUsesSeparateStableAndRCEndpoints(t *testing.T) {
	p := NewProvider(http.DefaultClient)
	stableURL, err := p.releaseURL(ChannelStable)
	if err != nil {
		t.Fatalf("stable release URL: %v", err)
	}
	if !strings.HasSuffix(stableURL, "/releases/latest") {
		t.Fatalf("stable release URL = %q, want releases/latest", stableURL)
	}
	rcURL, err := p.releaseURL(ChannelRC)
	if err != nil {
		t.Fatalf("RC release URL: %v", err)
	}
	if !strings.HasSuffix(rcURL, "/releases?per_page=100") {
		t.Fatalf("RC release URL = %q, want CatX release collection", rcURL)
	}
}

func TestProviderDiscoversHighestApprovedRCWithoutDowngrade(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := AssetBaseForTest(server.URL, Current)
		if r.URL.Path != "/api/releases" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `[
{"url":%q,"html_url":%q,"tag_name":"v0.2.0-rc.1","draft":false,"prerelease":true},
{"url":%q,"html_url":%q,"tag_name":"v0.2.0-rc.3","draft":true,"prerelease":true},
{"url":%q,"html_url":%q,"tag_name":"v0.2.0-rc.10","draft":false,"prerelease":true},
{"url":%q,"html_url":%q,"tag_name":"v0.2.0-rc.2","draft":false,"prerelease":true},
{"url":%q,"html_url":%q,"tag_name":"v0.2.0-rc.2+build","draft":false,"prerelease":true},
{"url":%q,"html_url":%q,"tag_name":"v0.2.0","draft":false,"prerelease":false},
{"url":"https://api.github.com/repos/other-owner/other-repo/releases/99","html_url":"https://github.com/other-owner/other-repo/releases/tag/v0.2.0-rc.99","tag_name":"v0.2.0-rc.99","draft":false,"prerelease":true}
]`,
			server.URL+"/api/releases/1", base+"/releases/tag/v0.2.0-rc.1",
			server.URL+"/api/releases/3", base+"/releases/tag/v0.2.0-rc.3",
			server.URL+"/api/releases/10", base+"/releases/tag/v0.2.0-rc.10",
			server.URL+"/api/releases/2", base+"/releases/tag/v0.2.0-rc.2",
			server.URL+"/api/releases/20", base+"/releases/tag/v0.2.0-rc.2+build",
			server.URL+"/api/releases/stable", base+"/releases/tag/v0.2.0")
	}))
	defer server.Close()

	base := AssetBaseForTest(server.URL, Current)
	provider := Provider{Identity: Current, Client: server.Client(), APIBaseURL: server.URL + "/api", WebBaseURL: base, AssetBaseURL: base}
	release, err := provider.Fetch(context.Background(), ChannelRC)
	if err != nil {
		t.Fatalf("discover RC release: %v", err)
	}
	if release.TagName != "v0.2.0-rc.10" {
		t.Fatalf("discovered RC = %q, want highest approved non-draft RC", release.TagName)
	}
}

func TestProviderRejectsRCWhenOnlyOlderReleaseLineExists(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := AssetBaseForTest(server.URL, Current)
		fmt.Fprintf(w, `[{"url":%q,"html_url":%q,"tag_name":"v0.1.0-rc.10","draft":false,"prerelease":true}]`,
			server.URL+"/api/releases/1", base+"/releases/tag/v0.1.0-rc.10")
	}))
	defer server.Close()
	base := AssetBaseForTest(server.URL, Current)
	provider := Provider{Identity: Current, Client: server.Client(), APIBaseURL: server.URL + "/api", WebBaseURL: base, AssetBaseURL: base}
	if _, err := provider.Fetch(context.Background(), ChannelRC); err == nil {
		t.Fatal("RC discovery accepted an older release line")
	}
}

func TestVerifySHA256RejectsMismatchAndWrongAssetName(t *testing.T) {
	payload := []byte("payload")
	digest := sha256.Sum256(payload)
	if err := VerifySHA256("asset", payload, []byte(fmt.Sprintf("%x  asset\n", digest))); err != nil {
		t.Fatalf("valid checksum rejected: %v", err)
	}
	if err := VerifySHA256("asset", payload, []byte(fmt.Sprintf("%x  other\n", digest))); err == nil {
		t.Fatal("checksum naming a different asset was accepted")
	}
	if err := VerifySHA256("asset", []byte("tampered"), []byte(fmt.Sprintf("%x  asset\n", digest))); err == nil {
		t.Fatal("checksum mismatch was accepted")
	}
}
