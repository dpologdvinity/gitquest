package levels

import (
	"fmt"
	"github.com/dpologdvinity/gitquest/internal/gitx"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type StepKind int

const (
	Write StepKind = iota
	Commit
	Branch
	Switch
	ResetHard
	Save
)

type Step struct {
	Kind                               StepKind
	Path, Content, Msg, Name, Ref, Key string
}

func W(path, content string) Step { return Step{Kind: Write, Path: path, Content: content} }
func C(msg string) Step           { return Step{Kind: Commit, Msg: msg} }
func B(name string) Step          { return Step{Kind: Branch, Name: name} }
func S(name string) Step          { return Step{Kind: Switch, Name: name} }
func RH(ref string) Step          { return Step{Kind: ResetHard, Ref: ref} }
func Sv(key, ref string) Step     { return Step{Kind: Save, Key: key, Ref: ref} }

type Level struct {
	N                  int
	Title, Story, Goal string
	Hints              [3]string
	Setup              []Step
	Check              func(*Ctx) (bool, string)
}

var registry []Level

func All() []Level {
	out := append([]Level(nil), registry...)
	sort.Slice(out, func(i, j int) bool { return out[i].N < out[j].N })
	return out
}
func ByN(n int) (Level, bool) {
	for _, l := range registry {
		if l.N == n {
			return l, true
		}
	}
	return Level{}, false
}
func ApplySteps(dir string, steps []Step) (map[string]string, error) {
	refs := map[string]string{}
	clock := gitx.SetupEpoch
	for _, s := range steps {
		g := gitx.Git{Dir: dir, Env: gitx.SetupEnv(clock)}
		var err error
		switch s.Kind {
		case Write:
			err = os.MkdirAll(filepath.Dir(filepath.Join(dir, s.Path)), 0755)
			if err == nil {
				err = os.WriteFile(filepath.Join(dir, s.Path), []byte(s.Content), 0644)
			}
		case Commit:
			if _, err = g.Run("add", "-A"); err == nil {
				_, err = g.Run("commit", "-q", "-m", s.Msg)
				clock += 60
			}
		case Branch:
			_, err = g.Run("branch", s.Name)
		case Switch:
			_, err = g.Run("switch", "-q", s.Name)
		case ResetHard:
			_, err = g.Run("reset", "-q", "--hard", s.Ref)
		case Save:
			var out string
			out, err = g.Run("rev-parse", s.Ref)
			refs[s.Key] = strings.TrimSpace(out)
		default:
			err = fmt.Errorf("unknown setup step %d", s.Kind)
		}
		if err != nil {
			return nil, err
		}
	}
	return refs, nil
}

type Ctx struct {
	Dir  string
	Refs map[string]string
}

func (c *Ctx) run(args ...string) string {
	g := gitx.Git{Dir: c.Dir, Env: gitx.CheckEnv()}
	out, _ := g.Run(append([]string{"-c", "core.excludesFile=" + os.DevNull}, args...)...)
	return out
}
func (c *Ctx) ok(args ...string) bool {
	g := gitx.Git{Dir: c.Dir, Env: gitx.CheckEnv()}
	ok, _ := g.Ok(append([]string{"-c", "core.excludesFile=" + os.DevNull}, args...)...)
	return ok
}
func (c *Ctx) Head() string { return strings.TrimSpace(c.run("symbolic-ref", "-q", "HEAD")) }
func (c *Ctx) Rev(name string) string {
	return strings.TrimSpace(c.run("rev-parse", "-q", "--verify", name+"^{commit}"))
}
func (c *Ctx) Has(rev, path string) bool    { return c.ok("rev-parse", "-q", "--verify", rev+":"+path) }
func (c *Ctx) Blob(rev, path string) string { return c.run("cat-file", "blob", rev+":"+path) }
func (c *Ctx) Parents(rev string) []string {
	fields := strings.Fields(c.run("rev-list", "--parents", "-n", "1", rev))
	if len(fields) > 0 {
		return fields[1:]
	}
	return nil
}
func (c *Ctx) Message(rev string) string {
	out := c.run("cat-file", "commit", rev)
	_, msg, _ := strings.Cut(out, "\n\n")
	return strings.TrimSpace(msg)
}
func (c *Ctx) Tree(rev string) string { return strings.TrimSpace(c.run("rev-parse", rev+"^{tree}")) }
func (c *Ctx) Count(args ...string) int {
	out := strings.TrimSpace(c.run(append([]string{"rev-list", "--count"}, args...)...))
	var n int
	_, _ = fmt.Sscan(out, &n)
	return n
}
func (c *Ctx) IsAncestor(a, b string) bool { return c.ok("merge-base", "--is-ancestor", a, b) }
func (c *Ctx) Ignored(path string) bool    { return c.ok("check-ignore", "-q", path) }
func (c *Ctx) Merging() bool               { return c.ok("rev-parse", "-q", "--verify", "MERGE_HEAD") }
func (c *Ctx) File(path string) (string, bool) {
	b, e := os.ReadFile(filepath.Join(c.Dir, path))
	return string(b), e == nil
}
func (c *Ctx) Ref(key string) string {
	if key == "master" || key == "main" {
		if ref := c.Refs[c.PrimaryBranch()]; ref != "" {
			return ref
		}
		other := "master"
		if c.PrimaryBranch() == "master" {
			other = "main"
		}
		return c.Refs[other]
	}
	return c.Refs[key]
}
func (c *Ctx) PrimaryBranch() string {
	if c.ok("show-ref", "--verify", "--quiet", "refs/heads/master") {
		return "master"
	}
	if c.ok("show-ref", "--verify", "--quiet", "refs/heads/main") || c.Head() == "refs/heads/main" {
		return "main"
	}
	return "master"
}
func (c *Ctx) OnMaster() bool { return c.Head() == "refs/heads/"+c.PrimaryBranch() }
