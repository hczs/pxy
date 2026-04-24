package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"pxy"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{"Usage:", "pxy init", "pxy on", "pxy off", "pxy status", "pxy test", "pxy list", "pxy config"} {
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
