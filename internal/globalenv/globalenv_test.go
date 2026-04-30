package globalenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteProfileSnippetIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(path, []byte("alias ll='ls -l'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{
		"http_proxy":  "http://127.0.0.1:7890",
		"HTTP_PROXY":  "http://127.0.0.1:7890",
		"all_proxy":   "socks5://127.0.0.1:7891",
		"ALL_PROXY":   "socks5://127.0.0.1:7891",
		"https_proxy": "http://127.0.0.1:7890",
		"HTTPS_PROXY": "http://127.0.0.1:7890",
	}
	if err := writeProfileSnippet(path, "bash", values); err != nil {
		t.Fatalf("writeProfileSnippet() error = %v", err)
	}
	if err := writeProfileSnippet(path, "bash", values); err != nil {
		t.Fatalf("writeProfileSnippet() second error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if count := strings.Count(content, startMarker); count != 1 {
		t.Fatalf("start marker count = %d, want 1:\n%s", count, content)
	}
	for _, want := range []string{
		"alias ll='ls -l'",
		"export http_proxy='http://127.0.0.1:7890'",
		"export ALL_PROXY='socks5://127.0.0.1:7891'",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("profile missing %q:\n%s", want, content)
		}
	}
}

func TestRemoveProfileSnippetPreservesOtherContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	values := map[string]string{"http_proxy": "http://127.0.0.1:7890"}
	if err := os.WriteFile(path, []byte("export EDITOR=vim\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeProfileSnippet(path, "zsh", values); err != nil {
		t.Fatalf("writeProfileSnippet() error = %v", err)
	}
	if err := removeProfileSnippet(path); err != nil {
		t.Fatalf("removeProfileSnippet() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "export EDITOR=vim") {
		t.Fatalf("profile content was not preserved:\n%s", content)
	}
	if strings.Contains(content, startMarker) || strings.Contains(content, "http_proxy") {
		t.Fatalf("pxy global block was not removed:\n%s", content)
	}
}

func TestReadProfileSnippet(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bashrc")
	values := map[string]string{
		"http_proxy": "http://127.0.0.1:7890",
		"all_proxy":  "socks5://127.0.0.1:7891",
	}
	if err := writeProfileSnippet(path, "bash", values); err != nil {
		t.Fatalf("writeProfileSnippet() error = %v", err)
	}
	got, err := readProfileSnippet(path)
	if err != nil {
		t.Fatalf("readProfileSnippet() error = %v", err)
	}
	for name, want := range values {
		if got[name] != want {
			t.Fatalf("value %s = %q, want %q; all=%v", name, got[name], want, got)
		}
	}
}

func TestPowerShellProfileSnippet(t *testing.T) {
	got, err := profileSnippet("powershell", map[string]string{
		"http_proxy": "http://127.0.0.1:7890",
	})
	if err != nil {
		t.Fatalf("profileSnippet() error = %v", err)
	}
	if !strings.Contains(got, "$env:http_proxy='http://127.0.0.1:7890'") {
		t.Fatalf("PowerShell snippet missing env assignment:\n%s", got)
	}
}
