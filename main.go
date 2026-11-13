// Command gitquest is an escape-room game that teaches git with real repositories.
package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/dpologdvinity/gitquest/internal/cli"
)

// version is set to the release version by GoReleaser with -ldflags -X.
var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.RealDeps(resolveVersion(version, debug.ReadBuildInfo))))
}

// resolveVersion keeps a version set at link time; otherwise, for builds such
// as `go install ...@v0.1.2`, it falls back to the module version embedded in
// the binary, and stays "dev" when there is none.
func resolveVersion(v string, read func() (*debug.BuildInfo, bool)) string {
	if v != "dev" {
		return v
	}
	if m := moduleVersion(read()); m != "" {
		return m
	}
	return v
}

// moduleVersion returns the main module's version without its "v" prefix, or
// "" when the build has no module version ("(devel)", reported for example by
// a build outside a git checkout or with -buildvcs=false). Inside a checkout,
// Go stamps a pseudo-version such as 0.1.3-0.20261111214722-23358f9f13f6+dirty.
func moduleVersion(info *debug.BuildInfo, ok bool) string {
	if !ok || info == nil || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return ""
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}
