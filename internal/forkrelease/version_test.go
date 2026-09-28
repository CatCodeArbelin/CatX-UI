package forkrelease

import "testing"

func TestCompareVersionStringsUsesSemverPrereleaseOrder(t *testing.T) {
	cases := []struct {
		name  string
		left  string
		right string
		want  int
	}{
		{"rc.1 before rc.2", "0.1.0-rc.1", "0.1.0-rc.2", -1},
		{"rc.2 before stable", "0.1.0-rc.2", "0.1.0", -1},
		{"stable after rc.1", "0.1.0", "0.1.0-rc.1", 1},
		{"numeric prerelease ordering", "1.0.0-rc.10", "1.0.0-rc.2", 1},
	}
	for _, tc := range cases {
		got, ok := CompareVersionStrings(tc.left, tc.right)
		if !ok || got != tc.want {
			t.Fatalf("%s: CompareVersionStrings(%q, %q) = %d, %v; want %d, true", tc.name, tc.left, tc.right, got, ok, tc.want)
		}
	}
}

func TestParseVersionRejectsMalformedPrereleaseTags(t *testing.T) {
	for _, input := range []string{
		"v0.1.0-rc.",
		"v0.1.0-rc.01",
		"v0.1.0-rc+build",
		"v0.1.0-rc_1",
		"v0.1.0-rc.1.2_3",
		"0.1.0-rc.1",
	} {
		if _, err := ParseReleaseTag(input); err == nil {
			t.Fatalf("ParseReleaseTag(%q) accepted malformed tag", input)
		}
	}
}
