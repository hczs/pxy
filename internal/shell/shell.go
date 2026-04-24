package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	marker    = "# pxy - proxy switcher"
	oldMarker = "# px - proxy switcher"
)

func Detect(env map[string]string, argv0 string) string {
	if env["PSModulePath"] != "" || env["PSVersionTable"] != "" {
		return "powershell"
	}
	if shell := env["SHELL"]; strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if shell := env["SHELL"]; strings.Contains(shell, "bash") {
		return "bash"
	}
	if strings.Contains(argv0, "zsh") {
		return "zsh"
	}
	if strings.Contains(argv0, "bash") {
		return "bash"
	}
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	return "bash"
}

func ProfilePath(shellName, home string) (string, error) {
	switch shellName {
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "powershell":
		if path := os.Getenv("PROFILE"); path != "" {
			return path, nil
		}
		return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

func FunctionSnippet(shellName, pxyBin string) (string, error) {
	switch shellName {
	case "bash", "zsh":
		// shell function 能修改当前 shell 环境；普通子进程不能修改父进程环境变量。
		return fmt.Sprintf(`%s
function pxy() {
  local pxy_bin="%s"
  case "$1" in
    on) eval "$("$pxy_bin" _on --shell %s)" ;;
    off) eval "$("$pxy_bin" _off --shell %s)" ;;
    *) "$pxy_bin" "$@" ;;
  esac
}
`, marker, escapeDouble(pxyBin), shellName, shellName), nil
	case "powershell":
		return fmt.Sprintf(`%s
function pxy {
  param(
    [Parameter(Position=0)]
    [string]$cmd,
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$rest
  )
  $pxyBin = '%s'
  switch ($cmd) {
    "on" { Invoke-Expression (& $pxyBin _on --shell powershell) }
    "off" { Invoke-Expression (& $pxyBin _off --shell powershell) }
    default {
      if ($cmd) { & $pxyBin $cmd @rest } else { & $pxyBin @rest }
    }
  }
}
`, marker, strings.ReplaceAll(pxyBin, "'", "''")), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

func Install(path, shellName, pxyBin string) error {
	snippet, err := FunctionSnippet(shellName, pxyBin)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create shell profile dir: %w", err)
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read shell profile: %w", err)
	}
	content := removeExistingSnippet(string(existing))
	if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + snippet
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write shell profile: %w", err)
	}
	return nil
}

func removeExistingSnippet(content string) string {
	start := len(content)
	for _, value := range []string{marker, oldMarker} {
		if index := strings.Index(content, value); index != -1 && index < start {
			start = index
		}
	}
	if start == len(content) {
		return content
	}
	return strings.TrimRight(content[:start], "\n")
}

func escapeDouble(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
