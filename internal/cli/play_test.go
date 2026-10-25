package cli

import (
	"fmt"
	"github.com/dpologdvinity/gitquest/internal/game"
	"github.com/dpologdvinity/gitquest/internal/levels"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStartFirstLevel(t *testing.T) {
	d, _, out, _ := testDeps(t)
	if Run([]string{"start"}, d) != 0 {
		t.Fatal(out)
	}
	l, _ := levels.ByN(1)
	want := "Level 1/12 · The Empty Logbook\n\n" + l.Story + "\n\nGoal: " + l.Goal + "\n\nYour level is ready:\n  cd ~/gq/level-01\nRun gitquest check when you think you're done, or gitquest hint if you're stuck.\n"
	if out.String() != want {
		t.Fatalf("got\n%s\nwant\n%s", out, want)
	}
}

func TestCheckReportsBrokenGitConfig(t *testing.T) {
	d, home, out, errb := testDeps(t)
	if Run([]string{"start"}, d) != 0 {
		t.Fatal(errb)
	}
	if err := os.WriteFile(filepath.Join(home, "gq", "level-01", ".git", "config"), []byte("[broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"check"}, d); code != 1 || out.Len() != 0 || !strings.Contains(errb.String(), "bad config line") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), errb.String())
	}
}
func TestLevelTwiceKeepsWork(t *testing.T) {
	d, home, out, _ := testDeps(t)
	Run([]string{"level", "3"}, d)
	work := filepath.Join(home, "gq", "level-03", "mine.txt")
	os.WriteFile(work, []byte("x"), 0644)
	out.Reset()
	if Run([]string{"level", "3"}, d) != 0 || !strings.Contains(out.String(), "Your level is already set up (gitquest reset rebuilds it):\n") {
		t.Fatal(out)
	}
	if _, e := os.Stat(work); e != nil {
		t.Fatal("player work lost")
	}
}
func TestLevelBadNumberAndForeignDir(t *testing.T) {
	d, home, _, errb := testDeps(t)
	for _, a := range []string{"0", "13", "x"} {
		errb.Reset()
		if Run([]string{"level", a}, d) != 2 || errb.String() != "gitquest: level must be a number from 1 to 12\n" {
			t.Fatal(a, errb)
		}
	}
	root := filepath.Join(home, "gq")
	stray := filepath.Join(root, "level-02")
	os.MkdirAll(stray, 0755)
	errb.Reset()
	if Run([]string{"level", "2"}, d) != 1 || errb.String() != "gitquest: "+root+" already exists and is not a gitquest folder (no .gitquest-root); set GITQUEST_ROOT to another path\n" {
		t.Fatal(errb)
	}
	os.WriteFile(filepath.Join(root, ".gitquest-root"), []byte("gitquest game root v1\n"), 0644)
	errb.Reset()
	if Run([]string{"level", "2"}, d) != 1 || errb.String() != "gitquest: refusing to touch "+stray+": it has no .gitquest-level marker; move or delete it yourself\n" {
		t.Fatal(errb)
	}
	if entries, _ := os.ReadDir(stray); len(entries) != 0 {
		t.Fatal("wrote into stray dir")
	}
}
func TestStatus(t *testing.T) {
	d, _, out, _ := testDeps(t)
	state := d.Getenv("GITQUEST_HOME")
	p, _, _ := game.LoadProgress(state, time.Time{})
	p.Level(1).SolvedAt = time.Unix(1, 0)
	p.Level(2).SolvedAt = time.Unix(1, 0)
	p.Current = 3
	p.Level(3).Hints = 2
	p.Level(4).Hints = 1
	game.SaveProgress(state, p)
	Run([]string{"status"}, d)
	lines := strings.Split(out.String(), "\n")
	want := []string{"gitquest · 2/12 solved", "", "✓  1 The Empty Logbook", "✓  2 A Fork in the Path", "▶  3 Catching Up  (2 hints used)", "   4 Two Captains  (1 hint used)", "   5 Typo in the Telegram"}
	if len(lines) != 15 || !reflect.DeepEqual(lines[:7], want) {
		t.Fatalf("%q", lines)
	}
}
func playerSh(t *testing.T, dir, script string) {
	t.Helper()
	home := t.TempDir()
	config := filepath.Join(t.TempDir(), "empty")
	os.WriteFile(config, nil, 0600)
	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+home, "GIT_CONFIG_GLOBAL="+config, "GIT_CONFIG_NOSYSTEM=1", "GIT_EDITOR=true")
	if b, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("%s: %s", e, b)
	}
}
func TestCheckFailThenPass(t *testing.T) {
	d, home, out, errb := testDeps(t)
	if Run([]string{"check"}, d) != 1 || errb.String() != "gitquest: no level in progress; run gitquest start\n" {
		t.Fatal(errb)
	}
	Run([]string{"start"}, d)
	out.Reset()
	if Run([]string{"check"}, d) != 1 || out.String() != "✗ Not solved yet: there are no commits yet\nStuck? Try gitquest hint\n" {
		t.Fatal(out)
	}
	dir := filepath.Join(home, "gq", "level-01")
	playerSh(t, dir, `git add logbook.txt && git commit -q -m "First entry"`)
	out.Reset()
	if Run([]string{"check"}, d) != 0 || out.String() != "✓ Level 1 solved: The Empty Logbook\nNext: gitquest level 2\n" {
		t.Fatal(out)
	}
}
func TestCheckUsesCwdLevel(t *testing.T) {
	d, home, out, _ := testDeps(t)
	Run([]string{"level", "2"}, d)
	Run([]string{"level", "1"}, d)
	dir := filepath.Join(home, "gq", "level-02")
	playerSh(t, dir, `git switch -q -c explore`)
	d.Getwd = func() (string, error) { return dir, nil }
	out.Reset()
	if Run([]string{"check"}, d) != 0 || !strings.HasPrefix(out.String(), "✓ Level 2 solved") {
		t.Fatal(out)
	}
}
func TestHints(t *testing.T) {
	d, _, out, _ := testDeps(t)
	Run([]string{"level", "8"}, d)
	l, _ := levels.ByN(8)
	for i := 0; i < 3; i++ {
		out.Reset()
		Run([]string{"hint"}, d)
		if out.String() != fmt.Sprintf("Hint %d/3: %s\n", i+1, l.Hints[i]) {
			t.Fatal(out)
		}
	}
	out.Reset()
	Run([]string{"hint"}, d)
	want := fmt.Sprintf("Hint 1/3: %s\nHint 2/3: %s\nHint 3/3: %s\nThat's every hint. You've got this!\n", l.Hints[0], l.Hints[1], l.Hints[2])
	if out.String() != want {
		t.Fatal(out)
	}
}
func TestResetRebuilds(t *testing.T) {
	d, home, out, _ := testDeps(t)
	Run([]string{"level", "6"}, d)
	dir := filepath.Join(home, "gq", "level-06")
	playerSh(t, dir, `git reset -q --hard HEAD~2`)
	out.Reset()
	if Run([]string{"reset"}, d) != 0 || !strings.HasPrefix(out.String(), "Level 6 rebuilt.\n\nLevel 6/12 · Undo the Spell\n") {
		t.Fatal(out)
	}
	out.Reset()
	if Run([]string{"check"}, d) != 1 || !strings.Contains(out.String(), "nothing has been undone yet") {
		t.Fatal(out)
	}
}
func TestResetRefusesUnmarked(t *testing.T) {
	d, home, _, errb := testDeps(t)
	Run([]string{"level", "4"}, d)
	dir := filepath.Join(home, "gq", "level-04")
	os.Remove(filepath.Join(dir, ".gitquest-level"))
	if Run([]string{"reset", "4"}, d) != 1 || errb.String() != "gitquest: refusing to touch "+dir+": it has no .gitquest-level marker; move or delete it yourself\n" {
		t.Fatal(errb)
	}
	if _, e := os.Stat(filepath.Join(dir, ".git")); e != nil {
		t.Fatal("deleted unmarked dir")
	}
}
func TestLastLevelMessage(t *testing.T) {
	d, home, out, _ := testDeps(t)
	Run([]string{"level", "12"}, d)
	playerSh(t, filepath.Join(home, "gq", "level-12"), `printf '*.log\ncache/\nsecrets.env\n' > .gitignore && git add .gitignore notes.md && git commit -q -m x`)
	out.Reset()
	Run([]string{"check"}, d)
	if out.String() != "✓ Level 12 solved: Secrets in the Attic\nThat was the last level. You escaped gitquest!\n" {
		t.Fatal(out)
	}
}
