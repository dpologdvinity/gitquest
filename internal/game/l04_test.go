package game

import (
	"github.com/dpologdvinity/gitquest/internal/levels"
	"path/filepath"
	"testing"
)

func init() {
	plays = append(plays, play{n: 4, fresh: "main is not a merge of both captains' orders yet", solve: `git merge captain-b; printf 'Course: west, then east\n' > orders.txt && git add orders.txt && git commit -q --no-edit`})
}
func TestMergeConflictReasons(t *testing.T) {
	for _, tc := range []struct{ script, want string }{{`git merge captain-b`, "the merge is still in progress: fix orders.txt, git add it, then git commit"}, {`git merge captain-b; printf 'Course: west\n' > orders.txt && git add orders.txt && git commit -q --no-edit`, "orders.txt should mention both west and east"}, {`git merge captain-b; git add orders.txt && git commit -q --no-edit`, "orders.txt still has conflict markers in it"}} {
		t.Run(tc.want, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			root := filepath.Join(t.TempDir(), "gq")
			l, _ := levels.ByN(4)
			dir, e := Build(root, l)
			if e != nil {
				t.Fatal(e)
			}
			sh(t, dir, tc.script)
			ok, reason, e := Check(root, l)
			if e != nil || ok || reason != tc.want {
				t.Fatal(ok, reason, e)
			}
		})
	}
}
