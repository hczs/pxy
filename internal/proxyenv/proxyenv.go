package proxyenv

import (
	"fmt"
	"strings"

	"github.com/hczs/px/internal/config"
)

var proxyVars = []string{"http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY", "all_proxy", "ALL_PROXY"}

func OnScript(shellName string, cfg config.Config) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
	}
	values := map[string]string{}
	if cfg.Proxy.HTTP.Enabled {
		url := fmt.Sprintf("http://%s:%d", cfg.Proxy.HTTP.Host, cfg.Proxy.HTTP.Port)
		values["http_proxy"] = url
		values["HTTP_PROXY"] = url
	}
	if cfg.Proxy.HTTPS.Enabled {
		url := fmt.Sprintf("http://%s:%d", cfg.Proxy.HTTPS.Host, cfg.Proxy.HTTPS.Port)
		values["https_proxy"] = url
		values["HTTPS_PROXY"] = url
	}
	if cfg.Proxy.SOCKS5.Enabled {
		url := fmt.Sprintf("socks5://%s:%d", cfg.Proxy.SOCKS5.Host, cfg.Proxy.SOCKS5.Port)
		values["all_proxy"] = url
		values["ALL_PROXY"] = url
	}

	switch shellName {
	case "bash", "zsh":
		return posixOn(values), nil
	case "powershell":
		return powerShellOn(values), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

func OffScript(shellName string) (string, error) {
	switch shellName {
	case "bash", "zsh":
		return posixOff(), nil
	case "powershell":
		return powerShellOff(), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

func posixOn(values map[string]string) string {
	var b strings.Builder
	for _, name := range proxyVars {
		// 开启前保存旧值，类似 Python dict 里先备份原来的 key，便于 px off 恢复。
		fmt.Fprintf(&b, "if [ \"${%s+x}\" ]; then __PX_OLD_%s=\"${%s}\"; __PX_OLD_%s_SET=1; else __PX_OLD_%s=''; __PX_OLD_%s_SET=0; fi\n", name, name, name, name, name, name)
	}
	for _, name := range proxyVars {
		if value, ok := values[name]; ok {
			fmt.Fprintf(&b, "export %s=%s\n", name, posixQuote(value))
		}
	}
	return b.String()
}

func posixOff() string {
	var b strings.Builder
	for _, name := range proxyVars {
		fmt.Fprintf(&b, "if [ \"${__PX_OLD_%s_SET:-0}\" = \"1\" ]; then export %s=\"${__PX_OLD_%s}\"; else unset %s; fi\n", name, name, name, name)
		fmt.Fprintf(&b, "unset __PX_OLD_%s __PX_OLD_%s_SET\n", name, name)
	}
	return b.String()
}

func powerShellOn(values map[string]string) string {
	var b strings.Builder
	for _, name := range proxyVars {
		fmt.Fprintf(&b, "if (Test-Path Env:%s) { $__PX_OLD_%s=$env:%s; $__PX_OLD_%s_SET=1 } else { $__PX_OLD_%s=''; $__PX_OLD_%s_SET=0 }\n", name, name, name, name, name, name)
	}
	for _, name := range proxyVars {
		if value, ok := values[name]; ok {
			fmt.Fprintf(&b, "$env:%s=%s\n", name, powerShellQuote(value))
		}
	}
	return b.String()
}

func powerShellOff() string {
	var b strings.Builder
	for _, name := range proxyVars {
		fmt.Fprintf(&b, "if (Test-Path variable:__PX_OLD_%s_SET) { if ($__PX_OLD_%s_SET -eq 1) { $env:%s=$__PX_OLD_%s } else { Remove-Item Env:%s -ErrorAction SilentlyContinue }; Remove-Variable __PX_OLD_%s,__PX_OLD_%s_SET -ErrorAction SilentlyContinue } else { Remove-Item Env:%s -ErrorAction SilentlyContinue }\n", name, name, name, name, name, name, name, name)
	}
	return b.String()
}

func posixQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func powerShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
