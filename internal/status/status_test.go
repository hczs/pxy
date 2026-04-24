package status

import (
	"strings"
	"testing"
)

func TestRenderEnabled(t *testing.T) {
	env := map[string]string{
		"http_proxy":  "http://127.0.0.1:7890",
		"https_proxy": "http://127.0.0.1:7890",
		"all_proxy":   "socks5://127.0.0.1:7891",
	}
	got := Render(env, "自动检测(Clash)")
	for _, want := range []string{"代理状态: 已开启", "HTTP:  http://127.0.0.1:7890", "SOCKS5: socks5://127.0.0.1:7891", "配置来源: 自动检测(Clash)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Render() missing %q:\n%s", want, got)
		}
	}
}
