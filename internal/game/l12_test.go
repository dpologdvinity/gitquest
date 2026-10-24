package game

import (
	"github.com/dpologdvinity/gitquest/internal/levels"
	"path/filepath"
	"testing"
)

func init() {
	plays = append(plays, play{n: 12, fresh: "there is no committed .gitignore yet", solve: `printf '*.log\ncache/\nsecrets.env\n' > .gitignore && git add .gitignore notes.md && git commit -q -m "Ignore the junk"`, wrong: `printf '*.log\ncache/\n' > .gitignore && git add .gitignore notes.md && git commit -q -m "Ignore some"`, wrongReason: "secrets.env is not ignored by your .gitignore"})
}
func TestCommittedJunkFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "gq")
	l, _ := levels.ByN(12)
	dir, e := Build(root, l)
	if e != nil {
		t.Fatal(e)
	}
	sh(t, dir, `printf '*.tmp\n' > .gitignore && git add -A && git commit -q -m all`)
	ok, reason, e := Check(root, l)
	if e != nil || ok || reason != "debug.log was committed; remove it with git rm --cached debug.log and commit again" {
		t.Fatal(ok, reason, e)
	}
}
