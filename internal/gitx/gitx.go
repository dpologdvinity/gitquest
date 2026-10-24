package gitx

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const SetupEpoch int64 = 1767258000

var ErrGitVersion = errors.New("needs git 2.32 or newer on your PATH")

type Git struct {
	Dir string
	Env []string
}

func baseEnv() []string {
	env := make([]string, 0)
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GIT_") {
			env = append(env, v)
		}
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "LC_ALL=C")
}
func SetupEnv(unix int64) []string {
	date := fmt.Sprintf("@%d +0000", unix)
	return append(baseEnv(), "GIT_AUTHOR_NAME=Gitquest Narrator", "GIT_AUTHOR_EMAIL=narrator@gitquest.invalid", "GIT_COMMITTER_NAME=Gitquest Narrator", "GIT_COMMITTER_EMAIL=narrator@gitquest.invalid", "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date, "TZ=UTC", "GIT_TERMINAL_PROMPT=0")
}
func CheckEnv() []string { return append(baseEnv(), "GIT_OPTIONAL_LOCKS=0") }
func (g Git) command(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir
	cmd.Env = g.Env
	return cmd
}
func (g Git) Run(args ...string) (string, error) {
	cmd := g.command(args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}
func (g Git) Ok(args ...string) (bool, error) {
	err := g.command(args...).Run()
	if err == nil {
		return true, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return false, nil
	}
	return false, err
}

var versionRE = regexp.MustCompile(`^git version ([0-9]+)\.([0-9]+)`)

func ParseVersion(s string) (major, minor int, ok bool) {
	m := versionRE.FindStringSubmatch(s)
	if m == nil {
		return 0, 0, false
	}
	a, e1 := strconv.Atoi(m[1])
	b, e2 := strconv.Atoi(m[2])
	return a, b, e1 == nil && e2 == nil
}
func VersionOK() error {
	out, err := exec.Command("git", "version").Output()
	if err != nil {
		return ErrGitVersion
	}
	a, b, ok := ParseVersion(string(out))
	if !ok || a < 2 || a == 2 && b < 32 {
		return ErrGitVersion
	}
	return nil
}
