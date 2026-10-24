package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testDeps(t *testing.T) (d Deps, home string, out, errb *bytes.Buffer) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	state := t.TempDir()
	env := map[string]string{"HOME": home, "GITQUEST_ROOT": filepath.Join(home, "gq"), "GITQUEST_HOME": state}
	out, errb = &bytes.Buffer{}, &bytes.Buffer{}
	return Deps{Version: "dev", Now: func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC) }, Getenv: func(k string) string { return env[k] }, Getwd: func() (string, error) { return home, nil }, Stdout: out, Stderr: errb}, home, out, errb
}
func TestVersionHelpUnknown(t *testing.T) {
	d, _, out, errb := testDeps(t)
	d.Version = "0.1.0"
	if Run([]string{"-V"}, d) != 0 || out.String() != "gitquest 0.1.0\n" {
		t.Fatal(out)
	}
	out.Reset()
	if Run([]string{"--help"}, d) != 0 || !strings.HasPrefix(out.String(), "gitquest - learn git by escaping puzzle rooms\n") {
		t.Fatal(out)
	}
	if Run([]string{"fly"}, d) != 2 || errb.String() != "gitquest: unknown command \"fly\" (see gitquest --help)\n" {
		t.Fatal(errb)
	}
}
