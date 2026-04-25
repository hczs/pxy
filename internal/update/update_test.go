package update

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
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
