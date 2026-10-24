package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Progress struct {
	Version int                    `json:"version"`
	Current int                    `json:"current"`
	Levels  map[string]*LevelState `json:"levels"`
}
type LevelState struct {
	SolvedAt time.Time `json:"solved_at"`
	Hints    int       `json:"hints"`
}

func StateDir(getenv func(string) string) (string, error) {
	if s := getenv("GITQUEST_HOME"); s != "" {
		return s, nil
	}
	if s := getenv("XDG_STATE_HOME"); filepath.IsAbs(s) {
		return filepath.Join(s, "gitquest"), nil
	}
	return filepath.Join(getenv("HOME"), ".local", "state", "gitquest"), nil
}
func freshProgress() *Progress { return &Progress{Version: 1, Levels: map[string]*LevelState{}} }
func LoadProgress(dir string, now time.Time) (*Progress, string, error) {
	path := filepath.Join(dir, "progress.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return freshProgress(), "", nil
	}
	if err != nil {
		return nil, "", err
	}
	p := freshProgress()
	err = json.Unmarshal(b, p)
	if err == nil {
		if p.Version != 1 || p.Current < 0 || p.Current > 12 {
			err = errors.New("invalid progress values")
		} else {
			for _, state := range p.Levels {
				if state == nil || state.Hints < 0 || state.Hints > 3 {
					err = errors.New("invalid progress values")
					break
				}
			}
		}
	}
	if err != nil {
		moved := fmt.Sprintf("%s.corrupt-%d", path, now.Unix())
		if e := os.Link(path, moved); e != nil {
			return nil, "", e
		}
		if e := os.Remove(path); e != nil {
			return nil, "", e
		}
		return freshProgress(), moved, nil
	}
	if p.Levels == nil {
		p.Levels = map[string]*LevelState{}
	}
	return p, "", nil
}
func SaveProgress(dir string, p *Progress) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".progress-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if err = f.Chmod(0600); err != nil {
		_ = f.Close()
		return err
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(dir, "progress.json"))
}
func (p *Progress) Level(n int) *LevelState {
	if p.Levels == nil {
		p.Levels = map[string]*LevelState{}
	}
	key := strconv.Itoa(n)
	if p.Levels[key] == nil {
		p.Levels[key] = &LevelState{}
	}
	return p.Levels[key]
}
func CurrentLevel(root, cwd string, p *Progress) int {
	resolve := func(s string) string {
		if v, e := filepath.EvalSymlinks(s); e == nil {
			return v
		}
		return filepath.Clean(s)
	}
	rel, err := filepath.Rel(resolve(root), resolve(cwd))
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		first := strings.Split(rel, string(os.PathSeparator))[0]
		if len(first) == 8 && strings.HasPrefix(first, "level-") {
			if n, e := strconv.Atoi(first[6:]); e == nil && n >= 1 && n <= 12 {
				return n
			}
		}
	}
	return p.Current
}
