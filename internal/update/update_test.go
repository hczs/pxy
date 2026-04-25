package update

import "testing"

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
