package game

import (
	"github.com/dpologdvinity/gitquest/internal/levels"
	"path/filepath"
	"testing"
)

func init() {
	plays = append(plays, play{n: 11, fresh: `there should be exactly one commit after "Write chapter 1"`, solve: `git reset -q --soft HEAD~2 && git commit -q -m "Write chapters 2 and 3"`, wrong: `git reset -q --hard HEAD~2 && git commit -q --allow-empty -m "Write chapters 2 and 3"`, wrongReason: "your uncommitted edit to ch1.txt is gone (git reset --hard throws working changes away); run gitquest reset and try --soft"})
}
func TestMixedResetCannotSquash(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "gq")
	l, _ := levels.ByN(11)
	dir, e := Build(root, l)
	if e != nil {
		t.Fatal(e)
	}
	sh(t, dir, `git reset -q --mixed HEAD~2 && git commit -q -m "Write chapters 2 and 3"`)
	ok, reason, e := Check(root, l)
	if e != nil || ok || reason != `there should be exactly one commit after "Write chapter 1"` {
		t.Fatal(ok, reason, e)
	}
}
