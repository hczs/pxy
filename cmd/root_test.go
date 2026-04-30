package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hczs/pxy/internal/globalenv"
)

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{"Usage:", "pxy init", "pxy on", "pxy off", "pxy status", "pxy test", "pxy list", "pxy global", "pxy config", "pxy version", "pxy update"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("help missing %q:\n%s", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy", "missing"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unknown command: missing") {
		t.Fatalf("stderr missing unknown command: %q", stderr.String())
	}
}

func TestInternalOnRequiresShell(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy", "_on"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "--shell is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestStatusCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy", "status"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "代理状态:") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestGlobalHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy", "global", "help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{"pxy global on", "pxy global off", "pxy global status"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("global help missing %q:\n%s", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestUnknownGlobalCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy", "global", "missing"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unknown global command: missing") {
		t.Fatalf("stderr missing unknown global command: %q", stderr.String())
	}
}

func TestRenderGlobalStatus(t *testing.T) {
	got := renderGlobalStatus(globalenv.Status{Values: map[string]string{
		"http_proxy": "http://127.0.0.1:7890",
		"all_proxy":  "socks5://127.0.0.1:7891",
	}})
	for _, want := range []string{"用户级永久代理状态: 已开启", "HTTP:  http://127.0.0.1:7890", "SOCKS5: socks5://127.0.0.1:7891"} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderGlobalStatus() missing %q:\n%s", want, got)
		}
	}
}
