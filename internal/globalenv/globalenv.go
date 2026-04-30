package globalenv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hczs/pxy/internal/config"
	"github.com/hczs/pxy/internal/proxyenv"
	"github.com/hczs/pxy/internal/shell"
)

const (
	startMarker = "# pxy - global proxy environment"
	endMarker   = "# pxy - end global proxy environment"
)

type Status struct {
	Values map[string]string
}

func (s Status) Enabled() bool {
	for _, value := range s.Values {
		if value != "" {
			return true
		}
	}
	return false
}

var (
	setUserEnv = defaultSetUserEnv
	getUserEnv = defaultGetUserEnv
)

func On(home, shellName string, cfg config.Config) (string, error) {
	values, err := proxyenv.Values(cfg)
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		for _, name := range proxyenv.VariableNames() {
			if value, ok := values[name]; ok {
				valueCopy := value
				if err := setUserEnv(name, &valueCopy); err != nil {
					return "", err
				}
			} else if err := setUserEnv(name, nil); err != nil {
				return "", err
			}
		}
		return "", nil
	}
	profile, err := shell.ProfilePath(shellName, home)
	if err != nil {
		return "", err
	}
	if err := writeProfileSnippet(profile, shellName, values); err != nil {
		return "", err
	}
	return profile, nil
}

func Off(home, shellName string) (string, error) {
	if runtime.GOOS == "windows" {
		for _, name := range proxyenv.VariableNames() {
			if err := setUserEnv(name, nil); err != nil {
				return "", err
			}
		}
		return "", nil
	}
	profile, err := shell.ProfilePath(shellName, home)
	if err != nil {
		return "", err
	}
	if err := removeProfileSnippet(profile); err != nil {
		return "", err
	}
	return profile, nil
}

func Check(home, shellName string) (Status, error) {
	if runtime.GOOS == "windows" {
		values := map[string]string{}
		for _, name := range proxyenv.VariableNames() {
			value, err := getUserEnv(name)
			if err != nil {
				return Status{}, err
			}
			if value != "" {
				values[name] = value
			}
		}
		return Status{Values: values}, nil
	}
	profile, err := shell.ProfilePath(shellName, home)
	if err != nil {
		return Status{}, err
	}
	values, err := readProfileSnippet(profile)
	if err != nil {
		return Status{}, err
	}
	return Status{Values: values}, nil
}

func writeProfileSnippet(path, shellName string, values map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create shell profile dir: %w", err)
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read shell profile: %w", err)
	}
	content := removeBlock(string(existing))
	if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	snippet, err := profileSnippet(shellName, values)
	if err != nil {
		return err
	}
	content += "\n" + snippet
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write shell profile: %w", err)
	}
	return nil
}

func removeProfileSnippet(path string) error {
	existing, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read shell profile: %w", err)
	}
	content := removeBlock(string(existing))
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write shell profile: %w", err)
	}
	return nil
}

func readProfileSnippet(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read shell profile: %w", err)
	}
	block := extractBlock(string(data))
	values := map[string]string{}
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export ") {
			keyValue := strings.TrimPrefix(line, "export ")
			name, quoted, ok := strings.Cut(keyValue, "=")
			if ok {
				values[name] = strings.Trim(quoted, "'")
			}
			continue
		}
		if strings.HasPrefix(line, "$env:") {
			keyValue := strings.TrimPrefix(line, "$env:")
			name, quoted, ok := strings.Cut(keyValue, "=")
			if ok {
				values[name] = strings.Trim(quoted, "'")
			}
		}
	}
	return values, nil
}

func profileSnippet(shellName string, values map[string]string) (string, error) {
	var b strings.Builder
	b.WriteString(startMarker + "\n")
	for _, name := range proxyenv.VariableNames() {
		if value, ok := values[name]; ok {
			switch shellName {
			case "bash", "zsh":
				fmt.Fprintf(&b, "export %s=%s\n", name, posixQuote(value))
			case "powershell":
				fmt.Fprintf(&b, "$env:%s=%s\n", name, powerShellQuote(value))
			default:
				return "", fmt.Errorf("unsupported shell: %s", shellName)
			}
		}
	}
	b.WriteString(endMarker + "\n")
	return b.String(), nil
}

func removeBlock(content string) string {
	start := strings.Index(content, startMarker)
	if start == -1 {
		return content
	}
	end := strings.Index(content[start:], endMarker)
	if end == -1 {
		return strings.TrimRight(content[:start], "\n")
	}
	end += start + len(endMarker)
	if end < len(content) && content[end] == '\r' {
		end++
	}
	if end < len(content) && content[end] == '\n' {
		end++
	}
	return strings.TrimRight(content[:start]+content[end:], "\n")
}

func extractBlock(content string) string {
	start := strings.Index(content, startMarker)
	if start == -1 {
		return ""
	}
	end := strings.Index(content[start:], endMarker)
	if end == -1 {
		return content[start:]
	}
	return content[start : start+end+len(endMarker)]
}

func defaultSetUserEnv(name string, value *string) error {
	valueScript := "$null"
	if value != nil {
		valueScript = powerShellQuote(*value)
	}
	script := fmt.Sprintf("[Environment]::SetEnvironmentVariable(%s, %s, 'User')", powerShellQuote(name), valueScript)
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set Windows user env %s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func defaultGetUserEnv(name string) (string, error) {
	script := fmt.Sprintf("[Console]::Out.Write([Environment]::GetEnvironmentVariable(%s, 'User'))", powerShellQuote(name))
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("get Windows user env %s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func posixQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func powerShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
