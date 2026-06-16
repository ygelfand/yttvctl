package main

import (
	"fmt"
	"os"

	"github.com/ygelfand/yttvctl/cmd"
)

// Populated via -ldflags by the Makefile / release build.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	if err := cmd.Root(cmd.BuildInfo{Version: Version, GitCommit: GitCommit, BuildDate: BuildDate}).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
