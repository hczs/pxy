package update

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	DefaultOwner = "hczs"
	DefaultRepo  = "pxy"
)

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

type Client struct {
	Owner          string
	Repo           string
	CurrentVersion string
	GOOS           string
	GOARCH         string
	ExecutablePath string
	HTTPClient     *http.Client
	APIBaseURL     string
}

type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	UpToDate       bool
	Asset          Asset
	ChecksumAsset  Asset
}

type releaseResponse struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

func CompareVersions(a, b string) int {
	ap := parseVersion(a)
	bp := parseVersion(b)
	for i := 0; i < len(ap); i++ {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(value string) [3]int {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	parts := strings.Split(value, ".")
	var out [3]int
	for i := 0; i < len(out) && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return [3]int{}
		}
		out[i] = n
	}
	return out
}

func FindArchiveAsset(assets []Asset, version, goos, goarch string) (Asset, error) {
	version = strings.TrimPrefix(version, "v")
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	want := fmt.Sprintf("pxy_%s_%s_%s%s", version, goos, goarch, ext)
	for _, asset := range assets {
		if asset.Name == want {
			return asset, nil
		}
	}
	return Asset{}, fmt.Errorf("find archive asset %s: not found", want)
}

func (c Client) Check(ctx context.Context) (CheckResult, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	baseURL := c.APIBaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	owner := c.Owner
	if owner == "" {
		owner = DefaultOwner
	}
	repo := c.Repo
	if repo == "" {
		repo = DefaultRepo
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/repos/%s/%s/releases/latest", strings.TrimRight(baseURL, "/"), owner, repo), nil)
	if err != nil {
		return CheckResult{}, fmt.Errorf("create latest release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "pxy")

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{}, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return CheckResult{}, fmt.Errorf("fetch latest release: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return CheckResult{}, fmt.Errorf("decode latest release: %w", err)
	}
	latest := strings.TrimPrefix(release.TagName, "v")
	result := CheckResult{
		CurrentVersion: c.CurrentVersion,
		LatestVersion:  latest,
		UpToDate:       CompareVersions(c.CurrentVersion, latest) >= 0,
	}
	if result.UpToDate {
		return result, nil
	}

	asset, err := FindArchiveAsset(release.Assets, latest, c.GOOS, c.GOARCH)
	if err != nil {
		return CheckResult{}, err
	}
	checksumAsset, err := findAsset(release.Assets, "checksums.txt")
	if err != nil {
		return CheckResult{}, err
	}
	result.Asset = asset
	result.ChecksumAsset = checksumAsset
	return result, nil
}

func findAsset(assets []Asset, name string) (Asset, error) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, nil
		}
	}
	return Asset{}, fmt.Errorf("find asset %s: not found", name)
}

func VerifyChecksum(data, checksums []byte, filename string) error {
	want, err := checksumFor(checksums, filename)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != want {
		return fmt.Errorf("verify checksum for %s: got %s, want %s", filename, got, want)
	}
	return nil
}

func checksumFor(checksums []byte, filename string) (string, error) {
	for _, line := range bytes.Split(checksums, []byte("\n")) {
		fields := strings.Fields(string(line))
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("find checksum for %s: not found", filename)
}
