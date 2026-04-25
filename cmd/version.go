package cmd

import (
	"fmt"
	"io"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func runVersion(stdout io.Writer) int {
	fmt.Fprintf(stdout, "pxy %s\ncommit: %s\nbuilt: %s\n", version, commit, date)
	return 0
}
