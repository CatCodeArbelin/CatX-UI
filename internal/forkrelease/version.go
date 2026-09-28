package forkrelease

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is the strict SemVer subset used by CatX release tags. Build
// metadata is intentionally not accepted: release identity must map to one
// immutable tag and one checksum set.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease []VersionIdentifier
}

type VersionIdentifier struct {
	Value   string
	Numeric bool
}

// ParseVersion accepts either the display form (0.1.0-rc.1) or a release tag
// with a leading v. It rejects malformed prerelease identifiers and numeric
// components with leading zeroes.
func ParseVersion(input string) (Version, error) {
	var version Version
	value := strings.TrimSpace(input)
	value = strings.TrimPrefix(value, "v")
	if value == "" || strings.Contains(value, "+") {
		return version, fmt.Errorf("invalid CatX version %q", input)
	}
	parts := strings.SplitN(value, "-", 2)
	core := strings.Split(parts[0], ".")
	if len(core) != 3 {
		return version, fmt.Errorf("invalid CatX version %q", input)
	}
	for index, item := range core {
		if !validNumericComponent(item) {
			return version, fmt.Errorf("invalid CatX version %q", input)
		}
		number, err := strconv.Atoi(item)
		if err != nil {
			return version, fmt.Errorf("invalid CatX version %q: %w", input, err)
		}
		switch index {
		case 0:
			version.Major = number
		case 1:
			version.Minor = number
		case 2:
			version.Patch = number
		}
	}
	if len(parts) == 1 {
		return version, nil
	}
	if parts[1] == "" {
		return Version{}, fmt.Errorf("invalid CatX prerelease %q", input)
	}
	for _, item := range strings.Split(parts[1], ".") {
		if !validPrereleaseIdentifier(item) {
			return Version{}, fmt.Errorf("invalid CatX prerelease %q", input)
		}
		version.Prerelease = append(version.Prerelease, VersionIdentifier{
			Value:   item,
			Numeric: isNumericIdentifier(item),
		})
	}
	return version, nil
}

func ParseReleaseTag(tag string) (Version, error) {
	if !strings.HasPrefix(tag, "v") {
		return Version{}, fmt.Errorf("release tag %q must start with v", tag)
	}
	return ParseVersion(tag)
}

func CompareVersionStrings(a, b string) (int, bool) {
	left, errLeft := ParseVersion(a)
	right, errRight := ParseVersion(b)
	if errLeft != nil || errRight != nil {
		return 0, false
	}
	return CompareVersions(left, right), true
}

func CompareVersions(a, b Version) int {
	if a.Major != b.Major {
		return compareInts(a.Major, b.Major)
	}
	if a.Minor != b.Minor {
		return compareInts(a.Minor, b.Minor)
	}
	if a.Patch != b.Patch {
		return compareInts(a.Patch, b.Patch)
	}
	if len(a.Prerelease) == 0 && len(b.Prerelease) == 0 {
		return 0
	}
	if len(a.Prerelease) == 0 {
		return 1
	}
	if len(b.Prerelease) == 0 {
		return -1
	}
	for index := 0; index < len(a.Prerelease) && index < len(b.Prerelease); index++ {
		left, right := a.Prerelease[index], b.Prerelease[index]
		if left.Numeric && right.Numeric {
			if len(left.Value) != len(right.Value) {
				return compareInts(len(left.Value), len(right.Value))
			}
		} else if left.Numeric != right.Numeric {
			if left.Numeric {
				return -1
			}
			return 1
		}
		if left.Value != right.Value {
			if left.Value < right.Value {
				return -1
			}
			return 1
		}
	}
	return compareInts(len(a.Prerelease), len(b.Prerelease))
}

func validNumericComponent(value string) bool {
	return value != "" && (value == "0" || value[0] != '0') && isDigits(value)
}

func validPrereleaseIdentifier(value string) bool {
	if value == "" || (isNumericIdentifier(value) && len(value) > 1 && value[0] == '0') {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && char != '-' {
			return false
		}
	}
	return true
}

func isNumericIdentifier(value string) bool { return value != "" && isDigits(value) }

func isDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func compareInts(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
