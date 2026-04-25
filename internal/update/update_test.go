package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal with v prefix", a: "v1.2.3", b: "1.2.3", want: 0},
		{name: "older patch", a: "1.2.3", b: "1.2.4", want: -1},
		{name: "newer minor", a: "1.3.0", b: "1.2.9", want: 1},
		{name: "dev is older than release", a: "dev", b: "1.0.0", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestFindArchiveAsset(t *testing.T) {
	assets := []Asset{
		{Name: "checksums.txt", DownloadURL: "https://example.test/checksums.txt"},
		{Name: "pxy_1.2.3_darwin_arm64.tar.gz", DownloadURL: "https://example.test/darwin"},
		{Name: "pxy_1.2.3_windows_amd64.zip", DownloadURL: "https://example.test/windows"},
	}

	got, err := FindArchiveAsset(assets, "1.2.3", "darwin", "arm64")
	if err != nil {
		t.Fatalf("FindArchiveAsset returned error: %v", err)
	}
	if got.Name != "pxy_1.2.3_darwin_arm64.tar.gz" {
		t.Fatalf("asset = %q, want darwin archive", got.Name)
	}
}

func TestFindArchiveAssetMissing(t *testing.T) {
	_, err := FindArchiveAsset([]Asset{{Name: "checksums.txt"}}, "1.2.3", "linux", "arm64")
	if err == nil {
		t.Fatal("FindArchiveAsset error = nil, want missing asset error")
	}
}

func TestClientCheckFindsLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/hczs/pxy/releases/latest" {
			t.Fatalf("path = %s, want latest release path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tag_name":"v1.2.3","assets":[{"name":"pxy_1.2.3_darwin_arm64.tar.gz","browser_download_url":"https://example.test/archive"},{"name":"checksums.txt","browser_download_url":"https://example.test/checksums"}]}`)
	}))
	t.Cleanup(server.Close)

	client := Client{
		Owner:          "hczs",
		Repo:           "pxy",
		CurrentVersion: "1.2.2",
		GOOS:           "darwin",
		GOARCH:         "arm64",
		HTTPClient:     server.Client(),
		APIBaseURL:     server.URL,
	}

	got, err := client.Check(context.Background())
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.UpToDate {
		t.Fatal("UpToDate = true, want false")
	}
	if got.LatestVersion != "1.2.3" {
		t.Fatalf("LatestVersion = %q, want 1.2.3", got.LatestVersion)
	}
	if got.Asset.Name != "pxy_1.2.3_darwin_arm64.tar.gz" {
		t.Fatalf("asset = %q, want release archive", got.Asset.Name)
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("archive bytes")
	sum := sha256.Sum256(data)
	checksums := fmt.Sprintf("%x  pxy_1.2.3_darwin_arm64.tar.gz\n", sum)

	if err := VerifyChecksum(data, []byte(checksums), "pxy_1.2.3_darwin_arm64.tar.gz"); err != nil {
		t.Fatalf("VerifyChecksum returned error: %v", err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	data := []byte("archive bytes")
	checksums := strings.Repeat("0", 64) + "  pxy_1.2.3_darwin_arm64.tar.gz\n"

	if err := VerifyChecksum(data, []byte(checksums), "pxy_1.2.3_darwin_arm64.tar.gz"); err == nil {
		t.Fatal("VerifyChecksum error = nil, want mismatch")
	}
}

func TestExtractTarGzBinary(t *testing.T) {
	archive := buildTarGz(t, "pxy", []byte("binary"))
	dir := t.TempDir()

	got, err := ExtractBinary(archive, "darwin", dir)
	if err != nil {
		t.Fatalf("ExtractBinary returned error: %v", err)
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read extracted binary: %v", err)
	}
	if string(data) != "binary" {
		t.Fatalf("binary = %q, want binary", data)
	}
}

func TestExtractZipBinary(t *testing.T) {
	archive := buildZip(t, "pxy.exe", []byte("binary"))
	dir := t.TempDir()

	got, err := ExtractBinary(archive, "windows", dir)
	if err != nil {
		t.Fatalf("ExtractBinary returned error: %v", err)
	}
	if filepath.Base(got) != "pxy.exe" {
		t.Fatalf("base = %q, want pxy.exe", filepath.Base(got))
	}
}

func TestUpdateReturnsManualPathOnWindows(t *testing.T) {
	archive := buildZip(t, "pxy.exe", []byte("binary"))
	sum := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%x  pxy_1.2.3_windows_amd64.zip\n", sum)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/hczs/pxy/releases/latest":
			fmt.Fprintf(w, `{"tag_name":"v1.2.3","assets":[{"name":"pxy_1.2.3_windows_amd64.zip","browser_download_url":"%s/archive.zip"},{"name":"checksums.txt","browser_download_url":"%s/checksums.txt"}]}`, serverURL(r), serverURL(r))
		case "/archive.zip":
			w.Write(archive)
		case "/checksums.txt":
			w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	exe := filepath.Join(t.TempDir(), "pxy.exe")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatalf("write exe: %v", err)
	}

	client := Client{
		Owner:          "hczs",
		Repo:           "pxy",
		CurrentVersion: "1.2.2",
		GOOS:           "windows",
		GOARCH:         "amd64",
		ExecutablePath: exe,
		HTTPClient:     server.Client(),
		APIBaseURL:     server.URL,
	}

	got, err := client.Update(context.Background())
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if got.Updated {
		t.Fatal("Updated = true, want false for Windows manual replacement")
	}
	if got.ManualPath == "" {
		t.Fatal("ManualPath is empty")
	}
	oldData, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read exe: %v", err)
	}
	if string(oldData) != "old" {
		t.Fatalf("exe changed on Windows: %q", oldData)
	}
}

func buildTarGz(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(data))}); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("write tar body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func buildZip(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func serverURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
