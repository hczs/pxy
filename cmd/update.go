package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/hczs/pxy/internal/update"
)

type updater interface {
	Check(context.Context) (update.CheckResult, error)
	Update(context.Context) (update.UpdateResult, error)
}

var newUpdater = func() (updater, error) {
	exe, err := executablePath()
	if err != nil {
		return nil, err
	}
	return update.Client{
		Owner:          update.DefaultOwner,
		Repo:           update.DefaultRepo,
		CurrentVersion: version,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
		ExecutablePath: exe,
		HTTPClient:     http.DefaultClient,
	}, nil
}

func runUpdate(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(stderr)
	checkOnly := fs.Bool("check", false, "check for updates without installing")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "unexpected argument: %s\n", fs.Arg(0))
		return 2
	}

	service, err := newUpdater()
	if err != nil {
		fmt.Fprintf(stderr, "update: %v\n", err)
		return 1
	}
	if *checkOnly {
		return runUpdateCheck(ctx, service, stdout, stderr)
	}
	return runUpdateInstall(ctx, service, stdout, stderr)
}

func runUpdateCheck(ctx context.Context, service updater, stdout, stderr io.Writer) int {
	result, err := service.Check(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "update: %v\n", err)
		return 1
	}
	if result.UpToDate {
		fmt.Fprintf(stdout, "pxy is up to date (%s)\n", result.CurrentVersion)
		return 0
	}
	fmt.Fprintf(stdout, "pxy %s is available (current %s)\n", result.LatestVersion, result.CurrentVersion)
	return 0
}

func runUpdateInstall(ctx context.Context, service updater, stdout, stderr io.Writer) int {
	result, err := service.Update(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "update: %v\n", err)
		return 1
	}
	if result.Check.UpToDate {
		fmt.Fprintf(stdout, "pxy is already up to date (%s)\n", result.Check.CurrentVersion)
		return 0
	}
	if result.ManualPath != "" {
		fmt.Fprintf(stdout, "downloaded pxy %s to %s; replace current executable manually on Windows\n", result.Check.LatestVersion, result.ManualPath)
		return 0
	}
	if result.Updated {
		fmt.Fprintf(stdout, "updated pxy from %s to %s\n", result.Check.CurrentVersion, result.Check.LatestVersion)
		return 0
	}
	fmt.Fprintf(stdout, "pxy update completed\n")
	return 0
}
