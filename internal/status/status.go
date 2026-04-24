package status

import (
	"fmt"
	"strings"
)

func Render(env map[string]string, source string) string {
	httpProxy := first(env, "http_proxy", "HTTP_PROXY")
	httpsProxy := first(env, "https_proxy", "HTTPS_PROXY")
	socksProxy := first(env, "all_proxy", "ALL_PROXY")
	var b strings.Builder
	if httpProxy == "" && httpsProxy == "" && socksProxy == "" {
		b.WriteString("代理状态: 已关闭\n")
	} else {
		b.WriteString("代理状态: 已开启\n")
		if httpProxy != "" {
			fmt.Fprintf(&b, "HTTP:  %s\n", httpProxy)
		}
		if httpsProxy != "" {
			fmt.Fprintf(&b, "HTTPS: %s\n", httpsProxy)
		}
		if socksProxy != "" {
			fmt.Fprintf(&b, "SOCKS5: %s\n", socksProxy)
		}
	}
	if source != "" {
		fmt.Fprintf(&b, "配置来源: %s\n", source)
	}
	return b.String()
}

func first(env map[string]string, names ...string) string {
	for _, name := range names {
		if value := env[name]; value != "" {
			return value
		}
	}
	return ""
}
