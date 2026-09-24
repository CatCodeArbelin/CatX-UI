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
			fmt.Fprintf(w, `{"url":%q,"html_url":%q,"tag_name":"v0.1.0","draft":false,"prerelease":false,"assets":[{"name":"catx-ui-update.sh","browser_download_url":%q,"size":%d},{"name":"catx-ui-update.sh.sha256","browser_download_url":%q,"size":84}]}`,
				server.URL+"/api/releases/1", base+"/releases/tag/v0.1.0", base+"/releases/download/v0.1.0/catx-ui-update.sh", len(payload), base+"/releases/download/v0.1.0/catx-ui-update.sh.sha256")
		case "/" + Current.RepositorySlug() + "/releases/download/v0.1.0/catx-ui-update.sh":
			_, _ = w.Write(payload)
		case "/" + Current.RepositorySlug() + "/releases/download/v0.1.0/catx-ui-update.sh.sha256":
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
	stable := &Release{APIURL: p.APIBaseURL + "/releases/1", HTMLURL: p.WebBaseURL + "/releases/tag/v0.1.0", TagName: "v0.1.0"}
	if err := p.ValidateRelease(stable, ChannelStable); err != nil {
		t.Fatalf("stable release rejected: %v", err)
	}
	stable.Prerelease = true
	if err := p.ValidateRelease(stable, ChannelStable); err == nil {
		t.Fatal("stable channel accepted a prerelease")
	}

	dev := &Release{APIURL: p.APIBaseURL + "/releases/2", HTMLURL: p.WebBaseURL + "/releases/tag/" + Current.DevReleaseTag, TagName: Current.DevReleaseTag, Prerelease: true}
	if err := p.ValidateRelease(dev, ChannelDev); err != nil {
		t.Fatalf("dev release rejected: %v", err)
	}
	dev.TagName = "nightly"
	if err := p.ValidateRelease(dev, ChannelDev); err == nil {
		t.Fatal("dev channel accepted an unexpected rolling tag")
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
