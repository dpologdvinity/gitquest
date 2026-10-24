package cli

import (
	"errors"
	"fmt"
	"github.com/dpologdvinity/gitquest/internal/game"
	"github.com/dpologdvinity/gitquest/internal/gitx"
	"github.com/dpologdvinity/gitquest/internal/levels"
	"io"
	"os"
	"strconv"
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
	if len(args) > 0 {
		switch args[0] {
		case "start", "level", "status":
			return runGame(args, d)
		}
	}
	if len(args) == 0 {
		fmt.Fprint(d.Stdout, usage)
		return 0
	}
	fmt.Fprintf(d.Stderr, "gitquest: unknown command %q (see gitquest --help)\n", args[0])
	return 2
}

func fail(d Deps, err error, code int) int { fmt.Fprintf(d.Stderr, "gitquest: %v\n", err); return code }
func parseLevel(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 12 {
		return 0, errors.New("level must be a number from 1 to 12")
	}
	return n, nil
}
func readyText(l levels.Level, dir string, existed bool) string {
	heading := "Your level is ready:"
	if existed {
		heading = "Your level is already set up (gitquest reset rebuilds it):"
	}
	return fmt.Sprintf("Level %d/12 · %s\n\n%s\n\nGoal: %s\n\n%s\n  cd %s\nRun gitquest check when you think you're done, or gitquest hint if you're stuck.\n", l.N, l.Title, l.Story, l.Goal, heading, dir)
}
func runGame(args []string, d Deps) int {
	cmd := args[0]
	if cmd == "level" && len(args) != 2 {
		return fail(d, errors.New("level must be a number from 1 to 12"), 2)
	}
	if cmd != "level" && len(args) != 1 {
		return fail(d, fmt.Errorf("unknown command %q (see gitquest --help)", cmd), 2)
	}
	n := 0
	if cmd == "level" {
		var err error
		n, err = parseLevel(args[1])
		if err != nil {
			return fail(d, err, 2)
		}
	}
	stateDir, err := game.StateDir(d.Getenv)
	if err != nil {
		return fail(d, err, 1)
	}
	p, moved, err := game.LoadProgress(stateDir, d.Now())
	if err != nil {
		return fail(d, err, 1)
	}
	if moved != "" {
		fmt.Fprintf(d.Stderr, "gitquest: progress file was unreadable; moved it to %s and started fresh\n", moved)
	}
	if cmd == "status" {
		solved := 0
		for _, l := range levels.All() {
			if !p.Level(l.N).SolvedAt.IsZero() {
				solved++
			}
		}
		fmt.Fprintf(d.Stdout, "gitquest · %d/12 solved\n\n", solved)
		for _, l := range levels.All() {
			state := p.Level(l.N)
			mark := " "
			if !state.SolvedAt.IsZero() {
				mark = "✓"
			} else if p.Current == l.N {
				mark = "▶"
			}
			line := fmt.Sprintf("%s %2d %s", mark, l.N, l.Title)
			if state.Hints > 0 {
				word := "hints"
				if state.Hints == 1 {
					word = "hint"
				}
				line += fmt.Sprintf("  (%d %s used)", state.Hints, word)
			}
			fmt.Fprintln(d.Stdout, line)
		}
		return 0
	}
	if cmd == "start" {
		for _, l := range levels.All() {
			if p.Level(l.N).SolvedAt.IsZero() {
				n = l.N
				break
			}
		}
		if n == 0 {
			fmt.Fprintln(d.Stdout, "You have solved all 12 levels. Replay any of them with gitquest level <N>.")
			return 0
		}
	}
	if err = gitx.VersionOK(); err != nil {
		return fail(d, err, 1)
	}
	root, err := game.Root(d.Getenv)
	if err != nil {
		code := 1
		if errors.Is(err, game.ErrRootRelative) {
			code = 2
		}
		return fail(d, err, code)
	}
	if err = game.EnsureRoot(root); err != nil {
		return fail(d, err, 1)
	}
	l, _ := levels.ByN(n)
	dir := game.LevelDir(root, n)
	existed := false
	if info, e := os.Lstat(dir); e == nil {
		existed = true
		if info.Mode()&os.ModeSymlink != 0 {
			return fail(d, &game.RefuseError{Dir: dir, Reason: "it is a symlink"}, 1)
		}
		m, e := game.ReadMarker(dir)
		if e != nil || m.Game != "gitquest" {
			return fail(d, &game.RefuseError{Dir: dir, Reason: "it has no .gitquest-level marker"}, 1)
		}
		if m.Level != n {
			return fail(d, &game.RefuseError{Dir: dir, Reason: fmt.Sprintf("its marker is for level %d", m.Level)}, 1)
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return fail(d, e, 1)
	}
	if !existed {
		if _, err = game.Build(root, l); err != nil {
			return fail(d, err, 1)
		}
	}
	p.Current = n
	if err = game.SaveProgress(stateDir, p); err != nil {
		return fail(d, err, 1)
	}
	fmt.Fprint(d.Stdout, readyText(l, game.DisplayPath(dir, d.Getenv("HOME")), existed))
	return 0
}
