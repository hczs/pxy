package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hczs/pxy/internal/update"
)

type fakeUpdater struct {
	checkResult  update.CheckResult
	updateResult update.UpdateResult
	err          error
	checkCalls   int
	updateCalls  int
}

func (f *fakeUpdater) Check(ctx context.Context) (update.CheckResult, error) {
	f.checkCalls++
	return f.checkResult, f.err
}

func (f *fakeUpdater) Update(ctx context.Context) (update.UpdateResult, error) {
	f.updateCalls++
	return f.updateResult, f.err
}

func TestUpdateCheckAvailable(t *testing.T) {
	fake := &fakeUpdater{
		checkResult: update.CheckResult{
			CurrentVersion: "1.0.0",
			LatestVersion:  "1.1.0",
			UpToDate:       false,
		},
	}
	restore := replaceUpdaterFactory(t, fake, nil)
	defer restore()

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"pxy", "update", "--check"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if fake.checkCalls != 1 || fake.updateCalls != 0 {
		t.Fatalf("calls check=%d update=%d, want check=1 update=0", fake.checkCalls, fake.updateCalls)
	}
	if !strings.Contains(stdout.String(), "pxy 1.1.0 is available") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestUpdateAlreadyCurrent(t *testing.T) {
	fake := &fakeUpdater{
		updateResult: update.UpdateResult{
			Check: update.CheckResult{
				CurrentVersion: "1.1.0",
				LatestVersion:  "1.1.0",
				UpToDate:       true,
			},
		},
	}
	restore := replaceUpdaterFactory(t, fake, nil)
	defer restore()

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"pxy", "update"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pxy is already up to date (1.1.0)") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestUpdateManualWindowsPath(t *testing.T) {
	fake := &fakeUpdater{
		updateResult: update.UpdateResult{
			Check: update.CheckResult{
				CurrentVersion: "1.0.0",
				LatestVersion:  "1.1.0",
			},
			ManualPath: `C:\Temp\pxy.exe`,
		},
	}
	restore := replaceUpdaterFactory(t, fake, nil)
	defer restore()

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"pxy", "update"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "replace current executable manually") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestUpdateError(t *testing.T) {
	fake := &fakeUpdater{err: errors.New("network down")}
	restore := replaceUpdaterFactory(t, fake, nil)
	defer restore()

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"pxy", "update", "--check"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "update: network down") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestUpdateInvalidArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"pxy", "update", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unexpected argument: extra") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func replaceUpdaterFactory(t *testing.T, fake updater, err error) func() {
	t.Helper()
	old := newUpdater
	newUpdater = func() (updater, error) {
		return fake, err
	}
	return func() {
		newUpdater = old
	}
}
