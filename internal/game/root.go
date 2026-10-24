package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const RootMarker = ".gitquest-root"
const RootMarkerContent = "gitquest game root v1\n"
const LevelMarker = ".gitquest-level"

var ErrRootRelative = errors.New("GITQUEST_ROOT must be an absolute path")

type Marker struct {
	Game  string            `json:"game"`
	Level int               `json:"level"`
	Refs  map[string]string `json:"refs"`
}
type RootError struct{ Root string }

func (e *RootError) Error() string {
	return fmt.Sprintf("%s already exists and is not a gitquest folder (no .gitquest-root); set GITQUEST_ROOT to another path", e.Root)
}

type RefuseError struct{ Dir, Reason string }

func (e *RefuseError) Error() string {
	return fmt.Sprintf("refusing to touch %s: %s; move or delete it yourself", e.Dir, e.Reason)
}
func Root(getenv func(string) string) (string, error) {
	if r := getenv("GITQUEST_ROOT"); r != "" {
		if !filepath.IsAbs(r) {
			return "", ErrRootRelative
		}
		return filepath.Clean(r), nil
	}
	return filepath.Join(getenv("HOME"), "gitquest"), nil
}
func LevelDir(root string, n int) string { return filepath.Join(root, fmt.Sprintf("level-%02d", n)) }
func EnsureRoot(root string) error {
	info, err := os.Lstat(root)
	if errors.Is(err, os.ErrNotExist) {
		if err = os.Mkdir(root, 0755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(root, RootMarker), []byte(RootMarkerContent), 0644)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return &RootError{root}
	}
	b, err := os.ReadFile(filepath.Join(root, RootMarker))
	if err == nil && string(b) == RootMarkerContent {
		return nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return os.WriteFile(filepath.Join(root, RootMarker), []byte(RootMarkerContent), 0644)
	}
	return &RootError{root}
}
func ReadMarker(dir string) (Marker, error) {
	var m Marker
	b, err := os.ReadFile(filepath.Join(dir, LevelMarker))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(b, &m)
	return m, err
}
func WriteMarker(dir string, m Marker) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, LevelMarker), append(b, '\n'), 0644)
}
func SafeRemove(root string, n int) error {
	if b, e := os.ReadFile(filepath.Join(root, RootMarker)); e != nil || string(b) != RootMarkerContent {
		return &RootError{root}
	}
	dir := LevelDir(root, n)
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	refuse := func(reason string) error { return &RefuseError{dir, reason} }
	if info.Mode()&os.ModeSymlink != 0 {
		return refuse("it is a symlink")
	}
	if !info.IsDir() {
		return refuse("it has no .gitquest-level marker")
	}
	parent, e1 := filepath.EvalSymlinks(filepath.Dir(dir))
	realRoot, e2 := filepath.EvalSymlinks(root)
	if e1 != nil || e2 != nil || parent != realRoot {
		return refuse("it is outside the game root")
	}
	m, e := ReadMarker(dir)
	if e != nil || m.Game != "gitquest" {
		return refuse("it has no .gitquest-level marker")
	}
	if m.Level != n {
		return refuse(fmt.Sprintf("its marker is for level %d", m.Level))
	}
	return os.RemoveAll(dir)
}
func DisplayPath(path, home string) string {
	if home != "" && strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
