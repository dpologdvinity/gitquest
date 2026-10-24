package game

import (
	"github.com/dpologdvinity/gitquest/internal/gitx"
	"github.com/dpologdvinity/gitquest/internal/levels"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var demo = levels.Level{N: 9, Title: "Demo", Setup: []levels.Step{levels.W("a.txt", "a\n"), levels.C("First"), levels.B("side"), levels.S("side"), levels.W("dir/b.txt", "b\n"), levels.C("Second"), levels.Sv("side", "HEAD"), levels.S("main"), levels.RH("main")}, Check: func(c *levels.Ctx) (bool, string) {
	if c.Rev("side") != c.Ref("side") {
		return false, "side moved"
	}
	return c.Has("side", "dir/b.txt") && !c.Has("main", "dir/b.txt") && c.OnMain(), "files"
}}

func TestBuildDeterministicAndIsolated(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	r1, r2 := filepath.Join(t.TempDir(), "gq"), filepath.Join(t.TempDir(), "gq")
	d1, e := Build(r1, demo)
	if e != nil {
		t.Fatal(e)
	}
	d2, e := Build(r2, demo)
	if e != nil {
		t.Fatal(e)
	}
	m1, _ := ReadMarker(d1)
	m2, _ := ReadMarker(d2)
	if m1.Level != 9 || m1.Refs["side"] == "" || m1.Refs["side"] != m2.Refs["side"] {
		t.Fatal(m1, m2)
	}
	g := gitx.Git{Dir: d1, Env: gitx.CheckEnv()}
	if msg, _ := g.Run("cat-file", "commit", m1.Refs["side"]); !strings.Contains(msg, "author Gitquest Narrator <narrator@gitquest.invalid> 1767258060 +0000") {
		t.Fatal(msg)
	}
	if out, _ := g.Run("config", "--local", "user.name"); out != "Adventurer\n" {
		t.Fatal(out)
	}
	if ok, _ := g.Ok("ls-files", "--error-unmatch", ".gitquest-level"); ok {
		t.Fatal("marker committed")
	}
	if ok, reason, e := Check(r1, demo); !ok || e != nil {
		t.Fatal(reason, e)
	}
	if _, e := Build(r1, demo); e == nil {
		t.Fatal("overbuilt")
	}
}
func TestPlayerGlobalConfigIgnored(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	hostile := filepath.Join(t.TempDir(), "gitconfig")
	os.WriteFile(hostile, []byte("[commit]\n\tgpgsign = true\n[merge]\n\tff = false\n[core]\n\texcludesFile = "+filepath.Join(t.TempDir(), "ign")+"\n"), 0644)
	t.Setenv("GIT_CONFIG_GLOBAL", hostile)
	root := filepath.Join(t.TempDir(), "gq")
	if _, e := Build(root, demo); e != nil {
		t.Fatal(e)
	}
	if ok, reason, _ := Check(root, demo); !ok {
		t.Fatal(reason)
	}
}
func TestCheckMissingLevel(t *testing.T) {
	root := filepath.Join(t.TempDir(), "gq")
	EnsureRoot(root)
	if _, _, e := Check(root, demo); e == nil || e.Error() != "level 9 is not set up; run gitquest level 9" {
		t.Fatal(e)
	}
}
