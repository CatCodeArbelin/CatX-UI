package forkrelease

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

type Channel string

const (
	ChannelStable Channel = "stable"
	ChannelDev    Channel = "dev"
)

var stableTagPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Release struct {
	APIURL          string  `json:"url"`
	HTMLURL         string  `json:"html_url"`
	TagName         string  `json:"tag_name"`
	Body            string  `json:"body"`
	TargetCommitish string  `json:"target_commitish"`
	Draft           bool    `json:"draft"`
	Prerelease      bool    `json:"prerelease"`
	Assets          []Asset `json:"assets"`
}

// Provider is a GitHub-release adapter with an explicit trust boundary. The
// alternate bases are useful for hermetic fake-server tests; production code
// uses NewProvider, whose bases are all pinned to CatX-UI's repository.
type Provider struct {
	Identity     Identity
	Client       *http.Client
	APIBaseURL   string
	WebBaseURL   string
	AssetBaseURL string
}

func NewProvider(client *http.Client) Provider {
	return Provider{
		Identity:     Current,
		Client:       client,
		APIBaseURL:   Current.APIRepositoryURL(),
		WebBaseURL:   Current.WebRepositoryURL(),
		AssetBaseURL: Current.WebRepositoryURL(),
	}
}

func (p Provider) releaseURL(channel Channel) (string, error) {
	switch channel {
	case ChannelStable:
		return strings.TrimRight(p.APIBaseURL, "/") + "/releases/latest", nil
	case ChannelDev:
		return strings.TrimRight(p.APIBaseURL, "/") + "/releases/tags/" + url.PathEscape(p.Identity.DevReleaseTag), nil
	default:
		return "", fmt.Errorf("unsupported release channel %q", channel)
	}
}

func (p Provider) Fetch(ctx context.Context, channel Channel) (*Release, error) {
	endpoint, err := p.releaseURL(channel)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release provider returned HTTP %d", resp.StatusCode)
	}
	var release Release
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 2<<20))
	if err := decoder.Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release metadata: %w", err)
	}
	if err := p.ValidateRelease(&release, channel); err != nil {
		return nil, err
	}
	return &release, nil
}

func (p Provider) ValidateRelease(release *Release, channel Channel) error {
	if release == nil {
		return fmt.Errorf("release metadata is nil")
	}
	if release.Draft {
		return fmt.Errorf("refusing draft release %q", release.TagName)
	}
	switch channel {
	case ChannelStable:
		if release.Prerelease {
			return fmt.Errorf("stable channel refuses prerelease %q", release.TagName)
		}
		if !stableTagPattern.MatchString(release.TagName) {
			return fmt.Errorf("stable release tag %q is not vMAJOR.MINOR.PATCH", release.TagName)
		}
	case ChannelDev:
		if release.TagName != p.Identity.DevReleaseTag {
			return fmt.Errorf("dev release tag %q does not match %q", release.TagName, p.Identity.DevReleaseTag)
		}
		if !release.Prerelease {
			return fmt.Errorf("dev release %q must be marked prerelease", release.TagName)
		}
	default:
		return fmt.Errorf("unsupported release channel %q", channel)
	}

	apiPrefix := strings.TrimRight(p.APIBaseURL, "/") + "/releases/"
	if !strings.HasPrefix(release.APIURL, apiPrefix) {
		return fmt.Errorf("release API URL is outside trusted repository: %q", release.APIURL)
	}
	wantHTML := strings.TrimRight(p.WebBaseURL, "/") + "/releases/tag/" + url.PathEscape(release.TagName)
	if release.HTMLURL != wantHTML {
		return fmt.Errorf("release page is outside trusted repository: %q", release.HTMLURL)
	}
	return nil
}

func (p Provider) FindAsset(release *Release, name string) (Asset, error) {
	if release == nil {
		return Asset{}, fmt.Errorf("release metadata is nil")
	}
	for _, asset := range release.Assets {
		if asset.Name != name {
			continue
		}
		want := strings.TrimRight(p.AssetBaseURL, "/") + "/releases/download/" +
			url.PathEscape(release.TagName) + "/" + url.PathEscape(name)
		if asset.BrowserDownloadURL != want {
			return Asset{}, fmt.Errorf("asset %q is outside trusted repository: %q", name, asset.BrowserDownloadURL)
		}
		if asset.Size <= 0 {
			return Asset{}, fmt.Errorf("asset %q has invalid size %d", name, asset.Size)
		}
		return asset, nil
	}
	return Asset{}, fmt.Errorf("release %q does not contain required asset %q", release.TagName, name)
}

func (p Provider) DownloadVerified(ctx context.Context, release *Release, assetName string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("maximum payload size must be positive")
	}
	asset, err := p.FindAsset(release, assetName)
	if err != nil {
		return nil, err
	}
	checksum, err := p.FindAsset(release, assetName+".sha256")
	if err != nil {
		return nil, fmt.Errorf("required checksum is missing: %w", err)
	}
	payload, err := p.download(ctx, asset, maxBytes)
	if err != nil {
		return nil, err
	}
	sums, err := p.download(ctx, checksum, 4096)
	if err != nil {
		return nil, err
	}
	if err := VerifySHA256(assetName, payload, sums); err != nil {
		return nil, err
	}
	return payload, nil
}

func (p Provider) download(ctx context.Context, asset Asset, maxBytes int64) ([]byte, error) {
	if asset.Size > maxBytes {
		return nil, fmt.Errorf("asset %q exceeds %d bytes", asset.Name, maxBytes)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, err
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %q: %w", asset.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %q returned HTTP %d", asset.Name, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", asset.Name, err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("asset %q exceeds %d bytes", asset.Name, maxBytes)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("asset %q is empty", asset.Name)
	}
	return data, nil
}

func VerifySHA256(assetName string, payload, checksumFile []byte) error {
	line := strings.TrimSpace(string(checksumFile))
	if strings.Contains(line, "\n") || strings.Contains(line, "\r") {
		return fmt.Errorf("checksum for %q must contain exactly one record", assetName)
	}
	fields := strings.Fields(line)
	if len(fields) != 2 || fields[1] != assetName {
		return fmt.Errorf("checksum record must name %q", assetName)
	}
	expected, err := hex.DecodeString(fields[0])
	if err != nil || len(expected) != sha256.Size {
		return fmt.Errorf("checksum for %q is not a valid SHA-256 digest", assetName)
	}
	actual := sha256.Sum256(payload)
	if !bytes.Equal(expected, actual[:]) {
		return fmt.Errorf("checksum mismatch for %q", assetName)
	}
	return nil
}

// AssetBaseForTest returns a repository-shaped base below serverURL. Keeping
// the repository slug in hermetic URLs lets tests exercise wrong-owner checks.
func AssetBaseForTest(serverURL string, identity Identity) string {
	return strings.TrimRight(serverURL, "/") + "/" + path.Join(identity.ReleaseOwner, identity.ReleaseRepository)
}
