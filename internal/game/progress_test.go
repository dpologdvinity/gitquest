package game

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func TestProgressRoundTripAndCorrupt(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "st")
	p, _, _ := LoadProgress(dir, t0)
	p.Current = 3
	p.Level(3).Hints = 2
	p.Level(1).SolvedAt = t0
	if e := SaveProgress(dir, p); e != nil {
		t.Fatal(e)
	}
	q, moved, e := LoadProgress(dir, t0)
	if e != nil || moved != "" || q.Current != 3 || q.Level(3).Hints != 2 || !q.Level(1).SolvedAt.Equal(t0) {
		t.Fatal(q, e)
	}
	os.WriteFile(filepath.Join(dir, "progress.json"), []byte("{{"), 0600)
	r, moved, _ := LoadProgress(dir, t0)
	if r.Current != 0 || moved != filepath.Join(dir, fmt.Sprintf("progress.json.corrupt-%d", t0.Unix())) {
		t.Fatal(r, moved)
	}
}
func TestCurrentLevelFromSubdirAndSymlink(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "gq")
	os.MkdirAll(filepath.Join(LevelDir(root, 7), "a", "b"), 0755)
	link := filepath.Join(base, "link")
	os.Symlink(root, link)
	p := &Progress{Current: 2}
	if n := CurrentLevel(root, filepath.Join(LevelDir(root, 7), "a", "b"), p); n != 7 {
		t.Fatal(n)
	}
	if n := CurrentLevel(root, filepath.Join(link, "level-07"), p); n != 7 {
		t.Fatal(n)
	}
	if n := CurrentLevel(link, LevelDir(root, 7), p); n != 7 {
		t.Fatal(n)
	}
	if n := CurrentLevel(root, base, p); n != 2 {
		t.Fatal(n)
	}
	if n := CurrentLevel(root, filepath.Join(root, "level-99"), &Progress{}); n != 0 {
		t.Fatal(n)
	}
}
