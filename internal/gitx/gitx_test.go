package gitx

import (
	"os"
	"strings"
	"testing"
)

func TestEnvsDropUserGitVars(t *testing.T) {
	t.Setenv("GIT_DIR", "/evil")
	t.Setenv("GIT_AUTHOR_NAME", "Me")
	env := strings.Join(SetupEnv(SetupEpoch+60), "\n")
	for _, w := range []string{"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_AUTHOR_NAME=Gitquest Narrator", "GIT_COMMITTER_EMAIL=narrator@gitquest.invalid", "GIT_AUTHOR_DATE=@1767258060 +0000", "GIT_COMMITTER_DATE=@1767258060 +0000", "LC_ALL=C", "TZ=UTC"} {
		if !strings.Contains(env, w) {
			t.Errorf("missing %s", w)
		}
	}
	if strings.Contains(env, "/evil") || strings.Contains(env, "GIT_AUTHOR_NAME=Me") {
		t.Fatal("user GIT_* leaked")
	}
	if c := strings.Join(CheckEnv(), "\n"); strings.Contains(c, "/evil") || !strings.Contains(c, "GIT_OPTIONAL_LOCKS=0") {
		t.Fatal(c)
	}
}
func TestRunAndOk(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	g := Git{Dir: t.TempDir(), Env: SetupEnv(SetupEpoch)}
	if _, e := g.Run("init", "-q", "-b", "main"); e != nil {
		t.Fatal(e)
	}
	if out, _ := g.Run("symbolic-ref", "HEAD"); out != "refs/heads/main\n" {
		t.Fatalf("%q", out)
	}
	if ok, e := g.Ok("rev-parse", "-q", "--verify", "HEAD"); ok || e != nil {
		t.Fatal(ok, e)
	}
	if _, e := g.Run("no-such-command"); e == nil {
		t.Fatal("want error")
	}
}
func TestParseVersion(t *testing.T) {
	for s, w := range map[string][2]int{"git version 2.55.0\n": {2, 55}, "git version 2.32.0.windows.1": {2, 32}, "git version 3.0": {3, 0}} {
		if a, b, ok := ParseVersion(s); !ok || a != w[0] || b != w[1] {
			t.Errorf("%q", s)
		}
	}
	if _, _, ok := ParseVersion("hub version"); ok {
		t.Fatal()
	}
}
