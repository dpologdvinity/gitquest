package game

import (
	"fmt"
	"github.com/dpologdvinity/gitquest/internal/gitx"
	"github.com/dpologdvinity/gitquest/internal/levels"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

type play struct {
	n                                int
	fresh, solve, wrong, wrongReason string
}

var plays []play

func playerEnv(t *testing.T) []string {
	t.Helper()
	home := t.TempDir()
	config := filepath.Join(t.TempDir(), "empty")
	os.WriteFile(config, nil, 0600)
	env := []string{}
	for _, v := range os.Environ() {
		if len(v) < 4 || v[:4] != "GIT_" {
			env = append(env, v)
		}
	}
	return append(env, "HOME="+home, "XDG_CONFIG_HOME="+home, "GIT_CONFIG_GLOBAL="+config, "GIT_CONFIG_NOSYSTEM=1", "GIT_EDITOR=true")
}
func sh(t *testing.T, dir, script string) {
	t.Helper()
	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = dir
	cmd.Env = playerEnv(t)
	cmd.Run()
}
func TestLevels(t *testing.T) {
	for _, p := range plays {
		l, ok := levels.ByN(p.n)
		if !ok {
			t.Fatalf("level %d not registered", p.n)
		}
		for _, tc := range []struct {
			name, script, want string
			solved             bool
		}{{"fresh", "", p.fresh, false}, {"solve", p.solve, "", true}, {"wrong", p.wrong, p.wrongReason, false}} {
			if tc.name == "wrong" && tc.script == "" {
				continue
			}
			t.Run(fmt.Sprintf("%02d/%s", p.n, tc.name), func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				root := filepath.Join(t.TempDir(), "gq")
				dir, e := Build(root, l)
				if e != nil {
					t.Fatal(e)
				}
				if tc.script != "" {
					sh(t, dir, tc.script)
				}
				got, reason, e := Check(root, l)
				if e != nil || got != tc.solved || !tc.solved && reason != tc.want {
					t.Fatalf("got %v %q %v, want %v %q", got, reason, e, tc.solved, tc.want)
				}
			})
		}
	}
}
func TestLevelsDeterministic(t *testing.T) {
	for _, p := range plays {
		t.Run(fmt.Sprint(p.n), func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			l, _ := levels.ByN(p.n)
			r1, r2 := filepath.Join(t.TempDir(), "gq"), filepath.Join(t.TempDir(), "gq")
			d1, e := Build(r1, l)
			if e != nil {
				t.Fatal(e)
			}
			d2, e := Build(r2, l)
			if e != nil {
				t.Fatal(e)
			}
			m1, _ := ReadMarker(d1)
			m2, _ := ReadMarker(d2)
			g1, g2 := gitx.Git{Dir: d1, Env: gitx.CheckEnv()}, gitx.Git{Dir: d2, Env: gitx.CheckEnv()}
			h1, _ := g1.Run("rev-parse", "HEAD")
			h2, _ := g2.Run("rev-parse", "HEAD")
			if !reflect.DeepEqual(m1.Refs, m2.Refs) || h1 != h2 {
				t.Fatal(m1, m2, h1, h2)
			}
		})
	}
}
