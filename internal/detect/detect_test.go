package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseClash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("mixed-port: 7890\nsocks-port: 7891\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result := ParseClash(path)
	if !result.Found || result.Name != "Clash" || result.Config.Proxy.HTTP.Port != 7890 || result.Config.Proxy.SOCKS5.Port != 7891 {
		t.Fatalf("ParseClash() = %+v", result)
	}
}

func TestParseSurge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "surge.conf")
	if err := os.WriteFile(path, []byte("[General]\nhttp-listen = 127.0.0.1:6152\nsocks5-listen = 127.0.0.1:6153\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result := ParseSurge(path)
	if !result.Found || result.Config.Proxy.HTTP.Port != 6152 || result.Config.Proxy.SOCKS5.Port != 6153 {
		t.Fatalf("ParseSurge() = %+v", result)
	}
}

func TestPickPreferred(t *testing.T) {
	results := []Result{
		{Name: "Surge", Found: true, Usable: true, Priority: 20},
		{Name: "Clash", Found: true, Usable: true, Priority: 10},
	}
	got, ok := PickPreferred(results)
	if !ok || got.Name != "Clash" {
		t.Fatalf("PickPreferred() = %+v, %v", got, ok)
	}
}
