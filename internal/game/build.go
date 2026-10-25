package game

import (
	"errors"
	"fmt"
	"github.com/dpologdvinity/gitquest/internal/gitx"
	"github.com/dpologdvinity/gitquest/internal/levels"
	"os"
	"path/filepath"
)

func Build(root string, l levels.Level) (dir string, err error) {
	if err = EnsureRoot(root); err != nil {
		return "", err
	}
	dir = LevelDir(root, l.N)
	if _, e := os.Lstat(dir); e == nil {
		return "", &RefuseError{dir, "it already exists"}
	} else if !errors.Is(e, os.ErrNotExist) {
		return "", e
	}
	if err = os.Mkdir(dir, 0755); err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = SafeRemove(root, l.N)
		}
	}()
	if err = WriteMarker(dir, Marker{Game: "gitquest", Level: l.N, Refs: map[string]string{}}); err != nil {
		return "", err
	}
	g := gitx.Git{Dir: dir, Env: gitx.SetupEnv(gitx.SetupEpoch)}
	if _, err = g.Run("init", "-q", "-b", "master"); err != nil {
		return "", err
	}
	f, e := os.OpenFile(filepath.Join(dir, ".git", "info", "exclude"), os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		return "", e
	}
	_, err = f.WriteString("\n.gitquest-level\n")
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}
	for _, kv := range [][2]string{{"user.name", "Adventurer"}, {"user.email", "adventurer@gitquest.invalid"}, {"commit.gpgsign", "false"}, {"tag.gpgsign", "false"}, {"merge.ff", "true"}, {"core.autocrlf", "false"}, {"core.hooksPath", filepath.Join(dir, ".git", "gitquest-no-hooks")}} {
		if _, err = g.Run("config", kv[0], kv[1]); err != nil {
			return "", err
		}
	}
	var refs map[string]string
	refs, err = levels.ApplySteps(dir, l.Setup)
	if err != nil {
		return "", err
	}
	err = WriteMarker(dir, Marker{Game: "gitquest", Level: l.N, Refs: refs})
	if err != nil {
		return "", err
	}
	return dir, nil
}
func Check(root string, l levels.Level) (bool, string, error) {
	dir := LevelDir(root, l.N)
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return false, "", fmt.Errorf("level %d is not set up; run gitquest level %d", l.N, l.N)
	}
	if err != nil {
		return false, "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, "", &RefuseError{dir, "it is a symlink"}
	}
	m, err := ReadMarker(dir)
	if err != nil || m.Game != "gitquest" {
		return false, "", &RefuseError{dir, "it has no .gitquest-level marker"}
	}
	if m.Level != l.N {
		return false, "", &RefuseError{dir, fmt.Sprintf("its marker is for level %d", m.Level)}
	}
	ctx := &levels.Ctx{Dir: dir, Refs: m.Refs}
	ok, reason := l.Check(ctx)
	if ctx.Err != nil {
		return false, "", ctx.Err
	}
	return ok, reason, nil
}
