package proxyenv

import (
	"strings"
	"testing"

	"github.com/hczs/pxy/internal/config"
)

func TestOnBashIncludesSnapshotAndExports(t *testing.T) {
	cfg := config.Config{Proxy: config.Proxy{
		HTTP:   config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: 7890},
		HTTPS:  config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: 7890},
		SOCKS5: config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: 7891},
	}}
	got, err := OnScript("bash", cfg)
	if err != nil {
		t.Fatalf("OnScript() error = %v", err)
	}
	for _, want := range []string{
		"__PX_OLD_http_proxy_SET=0",
		"export http_proxy='http://127.0.0.1:7890'",
		"export HTTPS_PROXY='http://127.0.0.1:7890'",
		"export all_proxy='socks5://127.0.0.1:7891'",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("script missing %q:\n%s", want, got)
		}
	}
}

func TestOffPowerShellRestoresSnapshot(t *testing.T) {
	got, err := OffScript("powershell")
	if err != nil {
		t.Fatalf("OffScript() error = %v", err)
	}
	for _, want := range []string{
		"if (Test-Path variable:__PX_OLD_http_proxy_SET)",
		"$env:http_proxy=$__PX_OLD_http_proxy",
		"Remove-Item Env:http_proxy -ErrorAction SilentlyContinue",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("script missing %q:\n%s", want, got)
		}
	}
}

func TestRejectsUnknownShell(t *testing.T) {
	_, err := OffScript("fish")
	if err == nil || !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("OffScript(fish) error = %v", err)
	}
}
