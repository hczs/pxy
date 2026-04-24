package interactive

import (
	"bytes"
	"strings"
	"testing"
)

func TestManualConfigDefaults(t *testing.T) {
	in := strings.NewReader("\n\n\n\n")
	var out bytes.Buffer

	cfg, err := ManualConfig(in, &out)
	if err != nil {
		t.Fatalf("ManualConfig() error = %v", err)
	}
	if cfg.Proxy.HTTP.Port != 7890 || cfg.Proxy.HTTPS.Port != 7890 || cfg.Proxy.SOCKS5.Port != 7891 {
		t.Fatalf("cfg = %+v", cfg)
	}
	if !strings.Contains(out.String(), "HTTP代理地址") {
		t.Fatalf("prompt output = %q", out.String())
	}
}

func TestConfirmNo(t *testing.T) {
	ok, err := Confirm(strings.NewReader("n\n"), &bytes.Buffer{}, "是否使用此配置？")
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if ok {
		t.Fatalf("Confirm() = true, want false")
	}
}
