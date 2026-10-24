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
func TestLoadProgressQuarantinesInvalidValues(t *testing.T) {
	for name, raw := range map[string]string{"current": `{"version":1,"current":99,"levels":{}}`, "hints": `{"version":1,"current":1,"levels":{"1":{"hints":-1}}}`} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "progress.json")
			if e := os.WriteFile(path, []byte(raw), 0600); e != nil {
				t.Fatal(e)
			}
			p, moved, e := LoadProgress(dir, t0)
			if e != nil || p == nil || p.Current != 0 || moved == "" {
				t.Fatal(p, moved, e)
			}
			b, e := os.ReadFile(moved)
			if e != nil || string(b) != raw {
				t.Fatal("original progress not preserved", e)
			}
		})
	}
}
func TestCorruptBackupIsNotOverwritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	backup := fmt.Sprintf("%s.corrupt-%d", path, t0.Unix())
	os.WriteFile(backup, []byte("older backup"), 0600)
	os.WriteFile(path, []byte("{{"), 0600)
	_, _, err := LoadProgress(dir, t0)
	if err == nil {
		t.Fatal("expected backup collision error")
	}
	if b, _ := os.ReadFile(backup); string(b) != "older backup" {
		t.Fatal("older backup overwritten")
	}
	if b, _ := os.ReadFile(path); string(b) != "{{" {
		t.Fatal("current corrupt file lost")
	}
}
