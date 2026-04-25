package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

type UpdateResult struct {
	Check      CheckResult
	Updated    bool
	ManualPath string
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

func (c Client) Update(ctx context.Context) (UpdateResult, error) {
	check, err := c.Check(ctx)
	if err != nil {
		return UpdateResult{}, err
	}
	if check.UpToDate {
		return UpdateResult{Check: check}, nil
	}

	archiveData, err := c.download(ctx, check.Asset.DownloadURL)
	if err != nil {
		return UpdateResult{}, err
	}
	checksumData, err := c.download(ctx, check.ChecksumAsset.DownloadURL)
	if err != nil {
		return UpdateResult{}, err
	}
	if err := VerifyChecksum(archiveData, checksumData, check.Asset.Name); err != nil {
		return UpdateResult{}, err
	}

	tmpDir, err := os.MkdirTemp("", "pxy-update-*")
	if err != nil {
		return UpdateResult{}, fmt.Errorf("create update temp dir: %w", err)
	}
	archivePath := filepath.Join(tmpDir, check.Asset.Name)
	if err := os.WriteFile(archivePath, archiveData, 0o600); err != nil {
		return UpdateResult{}, fmt.Errorf("write update archive: %w", err)
	}
	binaryPath, err := ExtractBinary(archiveData, c.GOOS, tmpDir)
	if err != nil {
		return UpdateResult{}, err
	}
	if c.GOOS == "windows" {
		return UpdateResult{Check: check, ManualPath: binaryPath}, nil
	}

	exe := c.ExecutablePath
	if exe == "" {
		exe, err = os.Executable()
		if err != nil {
			return UpdateResult{}, fmt.Errorf("resolve executable path: %w", err)
		}
	}
	if err := replaceExecutable(exe, binaryPath); err != nil {
		return UpdateResult{}, err
	}
	return UpdateResult{Check: check, Updated: true}, nil
}

func (c Client) download(ctx context.Context, url string) ([]byte, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("User-Agent", "pxy")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download %s: %w", url, err)
	}
	return data, nil
}

func ExtractBinary(archiveData []byte, goos, destDir string) (string, error) {
	if goos == "windows" {
		return extractZipBinary(archiveData, destDir)
	}
	return extractTarGzBinary(archiveData, destDir)
}

func extractTarGzBinary(archiveData []byte, destDir string) (string, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return "", fmt.Errorf("open tar.gz: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar.gz: %w", err)
		}
		if filepath.Base(header.Name) != "pxy" {
			continue
		}
		out := filepath.Join(destDir, "pxy")
		if err := writeExtractedFile(out, tr, 0o755); err != nil {
			return "", err
		}
		return out, nil
	}
	return "", fmt.Errorf("extract pxy from tar.gz: not found")
}

func extractZipBinary(archiveData []byte, destDir string) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	for _, file := range reader.File {
		if filepath.Base(file.Name) != "pxy.exe" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("open zip entry: %w", err)
		}
		defer rc.Close()
		out := filepath.Join(destDir, "pxy.exe")
		if err := writeExtractedFile(out, rc, 0o755); err != nil {
			return "", err
		}
		return out, nil
	}
	return "", fmt.Errorf("extract pxy.exe from zip: not found")
}

func writeExtractedFile(path string, src io.Reader, mode os.FileMode) error {
	dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create extracted binary: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("write extracted binary: %w", err)
	}
	return nil
}

func replaceExecutable(exePath, newPath string) error {
	info, err := os.Stat(exePath)
	if err != nil {
		return fmt.Errorf("stat executable: %w", err)
	}
	if err := os.Chmod(newPath, info.Mode()); err != nil {
		return fmt.Errorf("chmod new executable: %w", err)
	}

	targetDir := filepath.Dir(exePath)
	staged, err := os.CreateTemp(targetDir, ".pxy-update-*")
	if err != nil {
		return fmt.Errorf("stage new executable: %w", err)
	}
	stagedPath := staged.Name()
	defer os.Remove(stagedPath)

	src, err := os.Open(newPath)
	if err != nil {
		staged.Close()
		return fmt.Errorf("open new executable: %w", err)
	}
	if _, err := io.Copy(staged, src); err != nil {
		src.Close()
		staged.Close()
		return fmt.Errorf("copy new executable: %w", err)
	}
	if err := src.Close(); err != nil {
		staged.Close()
		return fmt.Errorf("close new executable: %w", err)
	}
	if err := staged.Chmod(info.Mode()); err != nil {
		staged.Close()
		return fmt.Errorf("chmod staged executable: %w", err)
	}
	if err := staged.Close(); err != nil {
		return fmt.Errorf("close staged executable: %w", err)
	}
	if err := os.Rename(stagedPath, exePath); err != nil {
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}
