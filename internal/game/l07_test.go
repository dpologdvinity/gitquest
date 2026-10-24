package game

import (
	"github.com/dpologdvinity/gitquest/internal/levels"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func init() {
	plays = append(plays, play{n: 7, fresh: "hotfix has no new commit yet", solve: `git stash -q && git switch -q hotfix && echo fixed > bridge.txt && git add bridge.txt && git commit -q -m "Fix the bridge" && git switch -q main && git stash pop -q`, wrong: `git stash -q && git switch -q hotfix && echo fixed > bridge.txt && git add bridge.txt && git commit -q -m "Fix the bridge" && git switch -q main`, wrongReason: "your plan B draft is not back in notes.txt yet"})
}
func TestHotfixNeedsStash(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "gq")
	l, _ := levels.ByN(7)
	dir, e := Build(root, l)
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command("git", "switch", "hotfix")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), playerEnv(t)...)
	if e := cmd.Run(); e == nil {
		t.Fatal("switch succeeded despite dirty notes")
	}
}
