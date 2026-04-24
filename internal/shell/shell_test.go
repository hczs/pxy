package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFunctionSnippetUsesAbsoluteBinary(t *testing.T) {
	got, err := FunctionSnippet("bash", "/tmp/pxy")
	if err != nil {
		t.Fatalf("FunctionSnippet() error = %v", err)
	}
	for _, want := range []string{`local pxy_bin="/tmp/pxy"`, `_on --shell bash`, `_off --shell bash`} {
		if !strings.Contains(got, want) {
			t.Fatalf("snippet missing %q:\n%s", want, got)
		}
	}
}

func TestInstallRemovesOldPxSnippet(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	old := "alias ll='ls -l'\n\n# px - proxy switcher\nfunction px() {}\n"
	if err := os.WriteFile(path, []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Install(path, "bash", "/tmp/pxy"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "# px - proxy switcher") || strings.Contains(content, "function px()") {
		t.Fatalf("old snippet was not removed:\n%s", content)
	}
	if !strings.Contains(content, "# pxy - proxy switcher") {
		t.Fatalf("new snippet missing:\n%s", content)
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(path, []byte("alias ll='ls -l'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Install(path, "bash", "/tmp/pxy"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if err := Install(path, "bash", "/tmp/pxy"); err != nil {
		t.Fatalf("Install() second error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(data), "# pxy - proxy switcher"); count != 1 {
		t.Fatalf("snippet count = %d, want 1:\n%s", count, string(data))
	}
}
