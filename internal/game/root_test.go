package game

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRootAndDisplay(t *testing.T) {
	g := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if r, _ := Root(g(map[string]string{"HOME": "/h"})); r != "/h/gitquest" {
		t.Fatal(r)
	}
	if r, _ := Root(g(map[string]string{"HOME": "/h", "GITQUEST_ROOT": "/x/gq"})); r != "/x/gq" {
		t.Fatal(r)
	}
	if _, e := Root(g(map[string]string{"GITQUEST_ROOT": "gq"})); !errors.Is(e, ErrRootRelative) {
		t.Fatal(e)
	}
	if LevelDir("/r", 3) != "/r/level-03" || DisplayPath("/h/gitquest/level-03", "/h") != "~/gitquest/level-03" || DisplayPath("/x/y", "/h") != "/x/y" {
		t.Fatal()
	}
}
func TestEnsureRoot(t *testing.T) {
	base := t.TempDir()
	fresh := filepath.Join(base, "new")
	if e := EnsureRoot(fresh); e != nil {
		t.Fatal(e)
	}
	if b, _ := os.ReadFile(filepath.Join(fresh, RootMarker)); string(b) != RootMarkerContent {
		t.Fatalf("%q", b)
	}
	if e := EnsureRoot(fresh); e != nil {
		t.Fatal(e)
	}
	empty := filepath.Join(base, "empty")
	os.Mkdir(empty, 0755)
	if e := EnsureRoot(empty); e != nil {
		t.Fatal(e)
	}
}
func TestEnsureRootRefusesForeignDir(t *testing.T) {
	proj := t.TempDir()
	os.WriteFile(filepath.Join(proj, "main.go"), []byte("package main"), 0644)
	var re *RootError
	if e := EnsureRoot(proj); !errors.As(e, &re) {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(proj, RootMarker)); e == nil {
		t.Fatal("wrote into foreign dir")
	}
	file := filepath.Join(t.TempDir(), "f")
	os.WriteFile(file, nil, 0644)
	if e := EnsureRoot(file); !errors.As(e, &re) {
		t.Fatal(e)
	}
}
func TestSafeRemoveRefusals(t *testing.T) {
	root := filepath.Join(t.TempDir(), "gq")
	EnsureRoot(root)
	mk := func(n int, m *Marker) string {
		d := LevelDir(root, n)
		os.MkdirAll(d, 0755)
		if m != nil {
			WriteMarker(d, *m)
		}
		return d
	}
	no := mk(1, nil)
	wrong := mk(2, &Marker{Game: "gitquest", Level: 5})
	outside := t.TempDir()
	precious := filepath.Join(outside, "precious")
	os.WriteFile(precious, []byte("x"), 0644)
	WriteMarker(outside, Marker{Game: "gitquest", Level: 3})
	os.Symlink(outside, LevelDir(root, 3))
	good := mk(4, &Marker{Game: "gitquest", Level: 4})
	for n, reason := range map[int]string{1: "it has no .gitquest-level marker", 2: "its marker is for level 5", 3: "it is a symlink"} {
		var re *RefuseError
		if e := SafeRemove(root, n); !errors.As(e, &re) || re.Reason != reason {
			t.Errorf("level %d: %v", n, e)
		}
	}
	for _, p := range []string{no, wrong, precious} {
		if _, e := os.Stat(p); e != nil {
			t.Fatalf("%s deleted", p)
		}
	}
	if e := SafeRemove(root, 4); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(good); !os.IsNotExist(e) {
		t.Fatal("not removed")
	}
	os.Remove(filepath.Join(root, RootMarker))
	mk(5, &Marker{Game: "gitquest", Level: 5})
	if e := SafeRemove(root, 5); e == nil {
		t.Fatal("must refuse")
	}
}
func TestEnsureRootDoesNotCreateParents(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Join(base, "unrequested")
	root := filepath.Join(parent, "gq")
	if e := EnsureRoot(root); e == nil {
		t.Fatal("created an unrequested parent")
	}
	if _, e := os.Lstat(parent); !os.IsNotExist(e) {
		t.Fatal("created parent outside game root")
	}
}
