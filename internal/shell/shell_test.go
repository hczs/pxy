package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFunctionSnippetUsesAbsoluteBinary(t *testing.T) {
	got, err := FunctionSnippet("bash", "/tmp/px")
	if err != nil {
		t.Fatalf("FunctionSnippet() error = %v", err)
	}
	for _, want := range []string{`local px_bin="/tmp/px"`, `_on --shell bash`, `_off --shell bash`} {
		if !strings.Contains(got, want) {
			t.Fatalf("snippet missing %q:\n%s", want, got)
		}
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(path, []byte("alias ll='ls -l'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Install(path, "bash", "/tmp/px"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if err := Install(path, "bash", "/tmp/px"); err != nil {
		t.Fatalf("Install() second error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(data), "# px - proxy switcher"); count != 1 {
		t.Fatalf("snippet count = %d, want 1:\n%s", count, string(data))
	}
}
