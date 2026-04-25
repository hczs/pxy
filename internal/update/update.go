package update

import (
	"fmt"
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
