package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hczs/pxy/internal/config"
	"github.com/hczs/pxy/internal/detect"
	"github.com/hczs/pxy/internal/globalenv"
	"github.com/hczs/pxy/internal/interactive"
	"github.com/hczs/pxy/internal/proxyenv"
	"github.com/hczs/pxy/internal/proxytest"
	"github.com/hczs/pxy/internal/shell"
	"github.com/hczs/pxy/internal/status"
)

const defaultTestURL = "https://ipwho.is/"

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	// Go 没有异常退出约定；CLI 常用 int 返回码表达成功或失败。
	if len(args) < 2 || args[1] == "-h" || args[1] == "--help" || args[1] == "help" {
		printHelp(stdout)
		return 0
	}

	switch args[1] {
	case "init":
		return runInit(args[2:], os.Stdin, stdout, stderr)
	case "config":
		return runConfig(os.Stdin, stdout, stderr)
	case "on":
		fmt.Fprintln(stderr, "请先执行 pxy init，然后在 shell 中使用 pxy on")
		return 1
	case "off":
		fmt.Fprintln(stderr, "请先执行 pxy init，然后在 shell 中使用 pxy off")
		return 1
	case "status":
		return runStatus(stdout)
	case "test":
		return runTest(ctx, stdout, stderr)
	case "list":
		return runList(stdout)
	case "global":
		return runGlobal(args[2:], stdout, stderr)
	case "version":
		return runVersion(stdout)
	case "update":
		return runUpdate(ctx, args[2:], stdout, stderr)
	case "_on":
		return runInternalOn(args[2:], stdout, stderr)
	case "_off":
		return runInternalOff(args[2:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[1])
		printHelp(stderr)
		return 2
	}
}

func runGlobal(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printGlobalHelp(stdout)
		return 0
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "get home dir: %v\n", err)
		return 1
	}
	shellName := shell.Detect(envMap(os.Environ()), filepath.Base(os.Args[0]))
	switch args[0] {
	case "on":
		cfg, err := config.Load(config.DefaultPath(home))
		if err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
		profile, err := globalenv.On(home, shellName, cfg)
		if err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
		if profile != "" {
			reload, err := shell.ReloadCommand(shellName, profile)
			if err != nil {
				fmt.Fprintf(stderr, "%v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "已写入用户级永久代理环境变量，新开的终端将生效；当前终端可执行 %s\n", reload)
			return 0
		}
		fmt.Fprintln(stdout, "已写入 Windows 用户级永久代理环境变量，新开的终端和应用将生效")
		return 0
	case "off":
		profile, err := globalenv.Off(home, shellName)
		if err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
		if profile != "" {
			reload, err := shell.ReloadCommand(shellName, profile)
			if err != nil {
				fmt.Fprintf(stderr, "%v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "已移除用户级永久代理环境变量配置，新开的终端将生效；当前终端可执行 %s\n", reload)
			return 0
		}
		fmt.Fprintln(stdout, "已移除 Windows 用户级永久代理环境变量，新开的终端和应用将生效")
		return 0
	case "status":
		result, err := globalenv.Check(home, shellName)
		if err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
		fmt.Fprint(stdout, renderGlobalStatus(result))
		return 0
	default:
		fmt.Fprintf(stderr, "unknown global command: %s\n", args[0])
		printGlobalHelp(stderr)
		return 2
	}
}

func runInit(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_ = args
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "get home dir: %v\n", err)
		return 1
	}
	shellName := shell.Detect(envMap(os.Environ()), filepath.Base(os.Args[0]))
	results := scanKnownConfigs(home)
	cfg, ok := detect.PickPreferred(results)
	var selected config.Config
	if ok {
		fmt.Fprintf(stdout, "检测到 %s 配置: %s\n", cfg.Name, cfg.Path)
		use, err := interactive.Confirm(stdin, stdout, "是否使用此配置？")
		if err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
		if use {
			selected = cfg.Config
		}
	}
	if selected.Proxy.Source == "" {
		var err error
		selected, err = interactive.ManualConfig(stdin, stdout)
		if err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
	}
	configPath := config.DefaultPath(home)
	if err := config.Save(configPath, selected); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	profile, err := shell.ProfilePath(shellName, home)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(stderr, "resolve pxy path: %v\n", err)
		return 1
	}
	if err := shell.Install(profile, shellName, exe); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	reload, err := shell.ReloadCommand(shellName, profile)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "pxy init 完成，请重启终端或执行 %s\n", reload)
	return 0
}

func runConfig(stdin io.Reader, stdout, stderr io.Writer) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "get home dir: %v\n", err)
		return 1
	}
	cfg, err := interactive.ManualConfig(stdin, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	if err := config.Save(config.DefaultPath(home), cfg); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "配置已保存")
	return 0
}

func runInternalOn(args []string, stdout, stderr io.Writer) int {
	shellName, ok := parseShellFlag(args, stderr)
	if !ok {
		return 2
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "get home dir: %v\n", err)
		return 1
	}
	cfg, err := config.Load(config.DefaultPath(home))
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	script, err := proxyenv.OnScript(shellName, cfg)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	fmt.Fprint(stdout, script)
	return 0
}

func runInternalOff(args []string, stdout, stderr io.Writer) int {
	shellName, ok := parseShellFlag(args, stderr)
	if !ok {
		return 2
	}
	script, err := proxyenv.OffScript(shellName)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	fmt.Fprint(stdout, script)
	return 0
}

func parseShellFlag(args []string, stderr io.Writer) (string, bool) {
	fs := flag.NewFlagSet("internal", flag.ContinueOnError)
	fs.SetOutput(stderr)
	shellName := fs.String("shell", "", "bash|zsh|powershell")
	if err := fs.Parse(args); err != nil {
		return "", false
	}
	if *shellName == "" {
		fmt.Fprintln(stderr, "--shell is required")
		return "", false
	}
	return *shellName, true
}

func runStatus(stdout io.Writer) int {
	source := ""
	if home, err := os.UserHomeDir(); err == nil {
		if cfg, err := config.Load(config.DefaultPath(home)); err == nil {
			source = cfg.Proxy.Source
		}
	}
	fmt.Fprint(stdout, status.Render(envMap(os.Environ()), source))
	return 0
}

func runTest(ctx context.Context, stdout, stderr io.Writer) int {
	result, err := proxytest.Run(ctx, defaultTestURL, &http.Client{})
	if err != nil {
		fmt.Fprintf(stderr, "代理测试失败，请确认已执行 pxy on: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "IP: %s\n国家: %s\n城市: %s\n", result.IP, result.Country, result.City)
	return 0
}

func runList(stdout io.Writer) int {
	home, _ := os.UserHomeDir()
	results := scanKnownConfigs(home)
	fmt.Fprintln(stdout, "检测到的代理软件：")
	for _, result := range results {
		if result.Found && result.Usable {
			fmt.Fprintf(stdout, "✓ %s (配置: %s)\n", result.Name, result.Path)
		} else if result.Found {
			fmt.Fprintf(stdout, "✗ %s (%s)\n", result.Name, result.Reason)
		} else {
			fmt.Fprintf(stdout, "✗ %s (未检测到)\n", result.Name)
		}
	}
	return 0
}

func scanKnownConfigs(home string) []detect.Result {
	candidates := []detect.Result{
		detect.ParseClash(filepath.Join(home, ".config", "clash", "config.yaml")),
		detect.ParseClash(filepath.Join(home, ".config", "clash-verge", "config.yaml")),
		detect.ParseSurge(filepath.Join(home, ".surge.conf")),
		detect.ParseV2Ray("v2rayA", "/etc/v2raya/config.json", 30),
		detect.ParseV2Ray("v2rayA", filepath.Join(home, ".config", "v2raya", "config.json"), 30),
	}
	if matches, err := filepath.Glob(filepath.Join(home, "Library", "Application Support", "Surge", "*.conf")); err == nil {
		for _, path := range matches {
			candidates = append(candidates, detect.ParseSurge(path))
		}
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		candidates = append(candidates, detect.ParseClash(filepath.Join(appData, "clash-verge", "config.yaml")))
		candidates = append(candidates, detect.ParseV2Ray("v2rayN", filepath.Join(appData, "v2rayN", "guiNConfig.json"), 40))
	} else {
		candidates = append(candidates, detect.ParseV2Ray("v2rayN", filepath.Join(home, "AppData", "Roaming", "v2rayN", "guiNConfig.json"), 40))
	}
	return candidates
}

func envMap(values []string) map[string]string {
	out := map[string]string{}
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if ok {
			out[key] = val
		}
	}
	return out
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  pxy init      Detect shell/proxy, save config, install shell function
  pxy on        Enable proxy in the current shell through the installed function
  pxy off       Restore or clear proxy variables in the current shell
  pxy status    Show current proxy environment
  pxy test      Test current proxy with https://ipwho.is/
  pxy list      List detected local proxy software
  pxy global    Manage user-level permanent proxy environment variables
  pxy config    Reconfigure proxy manually
  pxy version   Show build version
  pxy update    Check for and install a GitHub release update`)
}

func printGlobalHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  pxy global on       Enable user-level permanent proxy environment variables
  pxy global off      Remove user-level permanent proxy environment variables
  pxy global status   Show user-level permanent proxy environment status`)
}

func renderGlobalStatus(result globalenv.Status) string {
	var b strings.Builder
	if !result.Enabled() {
		b.WriteString("用户级永久代理状态: 已关闭\n")
		return b.String()
	}
	b.WriteString("用户级永久代理状态: 已开启\n")
	if value := firstGlobal(result.Values, "http_proxy", "HTTP_PROXY"); value != "" {
		fmt.Fprintf(&b, "HTTP:  %s\n", value)
	}
	if value := firstGlobal(result.Values, "https_proxy", "HTTPS_PROXY"); value != "" {
		fmt.Fprintf(&b, "HTTPS: %s\n", value)
	}
	if value := firstGlobal(result.Values, "all_proxy", "ALL_PROXY"); value != "" {
		fmt.Fprintf(&b, "SOCKS5: %s\n", value)
	}
	return b.String()
}

func firstGlobal(values map[string]string, names ...string) string {
	for _, name := range names {
		if value := values[name]; value != "" {
			return value
		}
	}
	return ""
}
