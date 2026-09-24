package main

import (
	"strings"
	"testing"
)

func TestReleaseInfoTextSeparatesTrustedIdentities(t *testing.T) {
	got := releaseInfoText()
	for _, want := range []string{
		"product=CatX-UI\n",
		"repository=CatCodeArbelin/CatX-UI\n",
		"fork_version=0.1.0\n",
		"upstream_base_version=3.8.5\n",
		"bundled_xray_version=26.9.9\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("release info %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "repository=MHSanaei/3x-ui") {
		t.Fatal("release identity unexpectedly trusts the official upstream repository")
	}
}
