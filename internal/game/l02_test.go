package game

import (
	"github.com/dpologdvinity/gitquest/internal/levels"
	"path/filepath"
	"testing"
)

func init() {
	plays = append(plays, play{n: 2, fresh: "there is no branch named explore", solve: `git switch -q -c explore`})
}
func TestExploreMustBeCurrent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "gq")
	l, _ := levels.ByN(2)
	dir, e := Build(root, l)
	if e != nil {
		t.Fatal(e)
	}
	sh(t, dir, "git branch explore")
	ok, reason, e := Check(root, l)
	if e != nil || ok || reason != "the explore branch exists, but you are not on it" {
		t.Fatal(ok, reason, e)
	}
}
