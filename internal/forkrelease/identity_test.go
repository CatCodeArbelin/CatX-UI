package forkrelease

import "testing"

func TestCurrentIdentityIsCatXOwned(t *testing.T) {
	if got, want := Current.RepositorySlug(), "CatCodeArbelin/CatX-UI"; got != want {
		t.Fatalf("repository slug = %q, want %q", got, want)
	}
	if Current.RepositorySlug() == "MHSanaei/3x-ui" {
		t.Fatal("fork updater identity must never resolve to the official upstream repository")
	}
	if ForkVersion() == "" || UpstreamBaseVersion() == "" {
		t.Fatal("fork and upstream-base versions must both be explicit")
	}
	if ForkVersion() == UpstreamBaseVersion() {
		t.Fatal("fork version and upstream-base version must remain separate values")
	}
	if Current.BundledXrayVersion == ForkVersion() || Current.BundledXrayVersion == UpstreamBaseVersion() {
		t.Fatal("bundled Xray version must remain separate from panel versions")
	}
}

func TestAssetNamesAreForkBranded(t *testing.T) {
	if got, want := Current.LinuxArchiveName("amd64"), "catx-ui-linux-amd64.tar.gz"; got != want {
		t.Fatalf("archive name = %q, want %q", got, want)
	}
	if got, want := Current.UpdateScriptAssetName(), "catx-ui-update.sh"; got != want {
		t.Fatalf("update script asset = %q, want %q", got, want)
	}
}
