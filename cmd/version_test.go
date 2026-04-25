package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	oldVersion := version
	oldCommit := commit
	oldDate := date
	t.Cleanup(func() {
		version = oldVersion
		commit = oldCommit
		date = oldDate
	})

	version = "1.2.3"
	commit = "abc123"
	date = "2026-04-25T00:00:00Z"

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"pxy", "version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{"pxy 1.2.3", "commit: abc123", "built: 2026-04-25T00:00:00Z"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
