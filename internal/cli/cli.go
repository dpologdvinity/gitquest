package cli

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Deps struct {
	Version        string
	Now            func() time.Time
	Getenv         func(string) string
	Getwd          func() (string, error)
	Stdout, Stderr io.Writer
}

func RealDeps(version string) Deps {
	return Deps{Version: version, Now: time.Now, Getenv: os.Getenv, Getwd: os.Getwd, Stdout: os.Stdout, Stderr: os.Stderr}
}

const usage = `gitquest - learn git by escaping puzzle rooms

usage:
  gitquest start          set up the first unsolved level
  gitquest level <N>      set up level N (1-12)
  gitquest check          check your solution
  gitquest hint           get a hint (3 per level)
  gitquest status         show your progress
  gitquest reset [N]      rebuild a level from scratch
  gitquest --version      print the version

Levels live in ~/gitquest (set GITQUEST_ROOT to change it).
Solve each level with real git commands inside its folder.
`

func Run(args []string, d Deps) int {
	if len(args) == 1 {
		switch args[0] {
		case "--version", "-V":
			fmt.Fprintf(d.Stdout, "gitquest %s\n", d.Version)
			return 0
		case "--help", "-h", "help":
			fmt.Fprint(d.Stdout, usage)
			return 0
		}
	}
	if len(args) == 0 {
		fmt.Fprint(d.Stdout, usage)
		return 0
	}
	fmt.Fprintf(d.Stderr, "gitquest: unknown command %q (see gitquest --help)\n", args[0])
	return 2
}
