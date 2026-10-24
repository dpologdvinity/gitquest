# gitquest v0.1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `gitquest`, a Go CLI escape-room game whose 12 levels are real git repos built deterministically in a marked game root and checked with git plumbing.

**Architecture:** `internal/gitx` runs the real `git` with isolated environments. `internal/levels` holds the data-defined levels (setup steps, texts, check funcs) and a plumbing-only `Ctx`. `internal/game` owns the game root, markers, safe deletion, building, checking and `progress.json`. `internal/cli` wires the commands.

**Tech Stack:** Go (`go 1.26` in go.mod), stdlib only, the user's `git` ≥ 2.32, GoReleaser v2, GitHub Actions.

**Spec:** `docs/SPEC.md`

## Global Constraints

- Module `github.com/dpologdvinity/gitquest`; `go.mod` line `go 1.26`; no third-party dependencies.
- License MIT, `Copyright (c) 2026 Kaitlyn Bassford`.
- Game root `$GITQUEST_ROOT` (absolute) else `$HOME/gitquest`; level dirs `level-01`…`level-12`; state `$GITQUEST_HOME`, else `$XDG_STATE_HOME/gitquest`, else `$HOME/.local/state/gitquest`.
- Gitquest creates and deletes only inside its root, and only dirs with a valid `.gitquest-level` marker (SPEC §5.3).
- Checks use only the plumbing listed in SPEC §6.3. Never parse `git status`, `git log` or other porcelain output.
- Setup is deterministic: env and clock exactly as SPEC §6.1–6.2 (clock starts at `1767258000`, +60 s after each commit).
- All level texts (story, goal, hints, reasons) are copied verbatim from SPEC §8.
- Tests never touch the real HOME, `~/gitquest` or real repos: `t.Setenv("HOME", t.TempDir())` in every test that builds levels; roots and state dirs under `t.TempDir()`.
- Errors `gitquest: <message>` on stderr; exit 1 runtime/check failure, 2 usage.
- Releases: linux/darwin × amd64/arm64, `CGO_ENABLED=0`.
- Conventional Commits. Never commit `CLAUDE.md`, `.claude/`, `AGENTS.md`, `.mcp.json`, `.env`.

## Review Focus

- `GITQUEST_ROOT` pointed at an existing project folder (or `$HOME` itself): gitquest must refuse and never write or delete there (test in Task 3: `TestEnsureRootRefusesForeignDir`).
- A level dir replaced by a symlink to somewhere else, or with its marker deleted, then `gitquest reset`: nothing outside the root may be deleted (test in Task 3: `TestSafeRemoveRefusals`).
- The player's own git config (aliases, `commit.gpgsign=true`, `merge.ff=false`, a global excludes file) must not break setup or change check results (test in Task 5: `TestPlayerGlobalConfigIgnored`).
- `gitquest check` run from a subdirectory of the level, or through a symlinked path to the root: it must still find the level (test in Task 4: `TestCurrentLevelFromSubdirAndSymlink`).
- Running `gitquest level N` twice: the second run must not rebuild over the player's work (test in Task 11: `TestLevelTwiceKeepsWork`).

---

### Task 1: Scaffold, version and help

**Files:**
- Create: `go.mod`, `main.go`, `LICENSE`, `.gitignore`, `internal/cli/cli.go`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.goreleaser.yaml`
- Test: `internal/cli/cli_test.go`

**Interfaces:**
- Produces:
  ```go
  package cli
  type Deps struct {
      Version string
      Now     func() time.Time
      Getenv  func(string) string
      Getwd   func() (string, error)
      Stdout, Stderr io.Writer
  }
  func Run(args []string, d Deps) int
  func RealDeps(version string) Deps
  ```
  Test helper (used by Tasks 11–12): `func testDeps(t *testing.T) (d Deps, home string, out, errb *bytes.Buffer)`: `home := t.TempDir()`, `t.Setenv("HOME", home)`, `Getenv` from a map with `HOME=home`, `GITQUEST_ROOT=home+"/gq"`, `GITQUEST_HOME=<another temp dir>`; `Getwd` returns `home`; `Now` returns `time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)`.

- [ ] **Step 1: Write the failing test**

```go
func TestVersionHelpUnknown(t *testing.T) {
    d, _, out, errb := testDeps(t); d.Version = "0.1.0"
    if Run([]string{"-V"}, d) != 0 || out.String() != "gitquest 0.1.0\n" { t.Fatal(out) }
    out.Reset()
    if Run([]string{"--help"}, d) != 0 || !strings.HasPrefix(out.String(), "gitquest - learn git by escaping puzzle rooms\n") { t.Fatal(out) }
    if Run([]string{"fly"}, d) != 2 || errb.String() != "gitquest: unknown command \"fly\" (see gitquest --help)\n" { t.Fatal(errb) }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/`
Expected: FAIL (build error: `undefined: Run`)

- [ ] **Step 3: Implement**

`go mod init github.com/dpologdvinity/gitquest`, set `go 1.26`. `main.go` with `var version = "dev"`. `Run` handles version, help (SPEC §4.1 verbatim) and unknown commands. `LICENSE` MIT `Copyright (c) 2026 Kaitlyn Bassford`; `.gitignore`: `dist/`, `.env`.
`ci.yml`: push + pull_request; matrix `[ubuntu-latest, macos-latest]`; `actions/checkout@v7`; `actions/setup-go@v7` with `go-version-file: go.mod`; `git --version`; `test -z "$(gofmt -l .)"`; `go vet ./...`; `go test ./...`.
`release.yml`: tags `v*`, `permissions: contents: write`, checkout `fetch-depth: 0`, setup-go, `goreleaser/goreleaser-action@v7` with `version: "~> v2"`, `args: release --clean`, `GITHUB_TOKEN`.
`.goreleaser.yaml`: `version: 2`, `project_name: gitquest`, `CGO_ENABLED=0`, `goos: [linux, darwin]`, `goarch: [amd64, arm64]`, `ldflags: ["-s -w -X main.version={{.Version}}"]`, archives `formats: [tar.gz]` with `LICENSE`, `README.md`; `checksum.name_template: checksums.txt`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/ && go vet ./... && test -z "$(gofmt -l .)"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod main.go LICENSE .gitignore internal/cli .github .goreleaser.yaml
git commit -m "chore: scaffold gitquest CLI with CI and release config"
```

---

### Task 2: Git runner and isolated environments

**Files:**
- Create: `internal/gitx/gitx.go`
- Test: `internal/gitx/gitx_test.go`

**Interfaces:**
- Produces:
  ```go
  const SetupEpoch int64 = 1767258000
  type Git struct { Dir string; Env []string }
  func SetupEnv(unix int64) []string            // SPEC §6.1 setup env
  func CheckEnv() []string                      // SPEC §6.1 check env
  func (g Git) Run(args ...string) (string, error) // raw stdout; error text includes trimmed stderr
  func (g Git) Ok(args ...string) (bool, error)    // exit 0 -> true; other exit code -> false, nil; cannot start -> error
  var ErrGitVersion = errors.New("needs git 2.32 or newer on your PATH")
  func ParseVersion(s string) (major, minor int, ok bool)
  func VersionOK() error                          // runs `git version`
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestEnvsDropUserGitVars(t *testing.T) {
    t.Setenv("GIT_DIR", "/evil"); t.Setenv("GIT_AUTHOR_NAME", "Me")
    env := strings.Join(SetupEnv(SetupEpoch+60), "\n")
    for _, want := range []string{"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Gitquest Narrator", "GIT_COMMITTER_EMAIL=narrator@gitquest.invalid", "GIT_AUTHOR_DATE=@1767258060 +0000", "GIT_COMMITTER_DATE=@1767258060 +0000", "LC_ALL=C", "TZ=UTC"} {
        if !strings.Contains(env, want) { t.Errorf("missing %s", want) }
    }
    if strings.Contains(env, "/evil") || strings.Contains(env, "GIT_AUTHOR_NAME=Me") { t.Fatal("user GIT_* leaked") }
    if c := strings.Join(CheckEnv(), "\n"); strings.Contains(c, "/evil") || !strings.Contains(c, "GIT_OPTIONAL_LOCKS=0") { t.Fatal(c) }
}
func TestRunAndOk(t *testing.T) {
    t.Setenv("HOME", t.TempDir())
    g := Git{Dir: t.TempDir(), Env: SetupEnv(SetupEpoch)}
    if _, err := g.Run("init", "-q", "-b", "master"); err != nil { t.Fatal(err) }
    if out, _ := g.Run("symbolic-ref", "HEAD"); out != "refs/heads/master\n" { t.Fatalf("%q", out) }
    if ok, err := g.Ok("rev-parse", "-q", "--verify", "HEAD"); ok || err != nil { t.Fatal(ok, err) }
    if _, err := g.Run("no-such-command"); err == nil { t.Fatal("want error") }
}
func TestParseVersion(t *testing.T) {
    for s, want := range map[string][2]int{"git version 2.55.0\n": {2, 55}, "git version 2.32.0.windows.1": {2, 32}, "git version 3.0": {3, 0}} {
        if a, b, ok := ParseVersion(s); !ok || a != want[0] || b != want[1] { t.Errorf("%q", s) }
    }
    if _, _, ok := ParseVersion("hub version"); ok { t.Fatal() }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/gitx/`
Expected: FAIL (build error: `undefined: SetupEnv`)

- [ ] **Step 3: Implement per SPEC §6.1 and the version rule in §6.3**

`VersionOK` returns `ErrGitVersion` if git is missing, unparseable, or below (2, 32).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/gitx/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/gitx
git commit -m "feat(gitx): run git with isolated setup and check environments"
```

---

### Task 3: Game root, markers and safe removal

**Files:**
- Create: `internal/game/root.go`
- Test: `internal/game/root_test.go`

**Interfaces:**
- Produces:
  ```go
  const RootMarker = ".gitquest-root"
  const RootMarkerContent = "gitquest game root v1\n"
  const LevelMarker = ".gitquest-level"
  type Marker struct { Game string `json:"game"`; Level int `json:"level"`; Refs map[string]string `json:"refs"` }
  type RootError struct{ Root string }       // Error(): "<root> already exists and is not a gitquest folder (no .gitquest-root); set GITQUEST_ROOT to another path"
  type RefuseError struct{ Dir, Reason string } // Error(): "refusing to touch <dir>: <reason>; move or delete it yourself"
  var ErrRootRelative = errors.New("GITQUEST_ROOT must be an absolute path")
  func Root(getenv func(string) string) (string, error)
  func LevelDir(root string, n int) string  // <root>/level-%02d
  func EnsureRoot(root string) error
  func ReadMarker(dir string) (Marker, error)
  func WriteMarker(dir string, m Marker) error
  func SafeRemove(root string, n int) error
  func DisplayPath(path, home string) string
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestRootAndDisplay(t *testing.T) {
    g := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
    if r, _ := Root(g(map[string]string{"HOME": "/h"})); r != "/h/gitquest" { t.Fatal(r) }
    if r, _ := Root(g(map[string]string{"HOME": "/h", "GITQUEST_ROOT": "/x/gq"})); r != "/x/gq" { t.Fatal(r) }
    if _, err := Root(g(map[string]string{"GITQUEST_ROOT": "gq"})); !errors.Is(err, ErrRootRelative) { t.Fatal(err) }
    if LevelDir("/r", 3) != "/r/level-03" || DisplayPath("/h/gitquest/level-03", "/h") != "~/gitquest/level-03" || DisplayPath("/x/y", "/h") != "/x/y" { t.Fatal() }
}
func TestEnsureRoot(t *testing.T) {
    base := t.TempDir()
    fresh := filepath.Join(base, "new")
    if err := EnsureRoot(fresh); err != nil { t.Fatal(err) }
    if b, _ := os.ReadFile(filepath.Join(fresh, RootMarker)); string(b) != RootMarkerContent { t.Fatalf("%q", b) }
    if err := EnsureRoot(fresh); err != nil { t.Fatal("second call must succeed") }
    empty := filepath.Join(base, "empty"); os.Mkdir(empty, 0o755)
    if err := EnsureRoot(empty); err != nil { t.Fatal(err) }
}
func TestEnsureRootRefusesForeignDir(t *testing.T) {
    proj := t.TempDir(); os.WriteFile(filepath.Join(proj, "main.go"), []byte("package main"), 0o644)
    var re *RootError
    if err := EnsureRoot(proj); !errors.As(err, &re) { t.Fatal(err) }
    if _, err := os.Stat(filepath.Join(proj, RootMarker)); err == nil { t.Fatal("wrote into a foreign dir") }
    file := filepath.Join(t.TempDir(), "f"); os.WriteFile(file, nil, 0o644)
    if err := EnsureRoot(file); !errors.As(err, &re) { t.Fatal(err) }
}
func TestSafeRemoveRefusals(t *testing.T) {
    root := filepath.Join(t.TempDir(), "gq"); EnsureRoot(root)
    mk := func(n int, m *Marker) string { d := LevelDir(root, n); os.MkdirAll(d, 0o755); if m != nil { WriteMarker(d, *m) }; return d }
    noMarker := mk(1, nil)
    wrong := mk(2, &Marker{Game: "gitquest", Level: 5})
    outside := t.TempDir(); os.WriteFile(filepath.Join(outside, "precious"), []byte("x"), 0o644)
    WriteMarker(outside, Marker{Game: "gitquest", Level: 3})
    os.Symlink(outside, LevelDir(root, 3))
    good := mk(4, &Marker{Game: "gitquest", Level: 4})
    var re *RefuseError
    for n, reason := range map[int]string{1: "it has no .gitquest-level marker", 2: "its marker is for level 5", 3: "it is a symlink"} {
        if err := SafeRemove(root, n); !errors.As(err, &re) || re.Reason != reason { t.Errorf("level %d: %v", n, err) }
    }
    for _, p := range []string{noMarker, wrong, filepath.Join(outside, "precious")} { if _, err := os.Stat(p); err != nil { t.Fatalf("%s deleted", p) } }
    if err := SafeRemove(root, 4); err != nil { t.Fatal(err) }
    if _, err := os.Stat(good); !os.IsNotExist(err) { t.Fatal("not removed") }
    os.Remove(filepath.Join(root, RootMarker)); mk(5, &Marker{Game: "gitquest", Level: 5})
    if err := SafeRemove(root, 5); err == nil { t.Fatal("must refuse when the root marker is gone") }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/game/`
Expected: FAIL (build error: `undefined: Root`)

- [ ] **Step 3: Implement per SPEC §5.1–5.3**

`SafeRemove` checks in this order: root marker (else `RootError`), `Lstat` symlink → `it is a symlink`, not a dir → `it has no .gitquest-level marker`, parent mismatch after `EvalSymlinks` → `it is outside the game root`, marker missing/unparseable/`game != "gitquest"` → `it has no .gitquest-level marker`, level mismatch → `its marker is for level <M>`. A missing level dir is not an error (returns nil).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/game/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/game
git commit -m "feat(game): add marked game root and refuse-by-default level removal"
```

---

### Task 4: Progress and current level

**Files:**
- Create: `internal/game/progress.go`
- Test: `internal/game/progress_test.go`

**Interfaces:**
- Produces:
  ```go
  type Progress struct { Version int; Current int; Levels map[string]*LevelState } // JSON tags per SPEC §5.4
  type LevelState struct { SolvedAt time.Time; Hints int }
  func StateDir(getenv func(string) string) (string, error)
  func LoadProgress(dir string, now time.Time) (p *Progress, movedTo string, err error) // missing -> empty; corrupt -> moved
  func SaveProgress(dir string, p *Progress) error  // atomic, 0600, dir 0700
  func (p *Progress) Level(n int) *LevelState       // get-or-create Levels[strconv.Itoa(n)]
  func CurrentLevel(root, cwd string, p *Progress) int
  ```

- [ ] **Step 1: Write the failing tests**

```go
var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func TestProgressRoundTripAndCorrupt(t *testing.T) {
    dir := filepath.Join(t.TempDir(), "st")
    p, _, _ := LoadProgress(dir, t0)
    p.Current = 3; p.Level(3).Hints = 2; p.Level(1).SolvedAt = t0
    if err := SaveProgress(dir, p); err != nil { t.Fatal(err) }
    q, moved, err := LoadProgress(dir, t0)
    if err != nil || moved != "" || q.Current != 3 || q.Level(3).Hints != 2 || !q.Level(1).SolvedAt.Equal(t0) { t.Fatal(q, err) }
    os.WriteFile(filepath.Join(dir, "progress.json"), []byte("{{"), 0o600)
    r, moved, _ := LoadProgress(dir, t0)
    if r.Current != 0 || moved != filepath.Join(dir, fmt.Sprintf("progress.json.corrupt-%d", t0.Unix())) { t.Fatal(r, moved) }
}
func TestCurrentLevelFromSubdirAndSymlink(t *testing.T) {
    base := t.TempDir(); root := filepath.Join(base, "gq")
    os.MkdirAll(filepath.Join(LevelDir(root, 7), "a", "b"), 0o755)
    link := filepath.Join(base, "link"); os.Symlink(root, link)
    p := &Progress{Current: 2}
    if n := CurrentLevel(root, filepath.Join(LevelDir(root, 7), "a", "b"), p); n != 7 { t.Fatal(n) }
    if n := CurrentLevel(root, filepath.Join(link, "level-07"), p); n != 7 { t.Fatal(n) }
    if n := CurrentLevel(link, LevelDir(root, 7), p); n != 7 { t.Fatal(n) }
    if n := CurrentLevel(root, base, p); n != 2 { t.Fatal(n) }
    if n := CurrentLevel(root, filepath.Join(root, "level-99"), &Progress{}); n != 0 { t.Fatal(n) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/game/ -run 'Progress|CurrentLevel'`
Expected: FAIL (build error: `undefined: LoadProgress`)

- [ ] **Step 3: Implement per SPEC §5.1, §5.4, §5.5**

`CurrentLevel`: resolve `root` and `cwd` with `filepath.EvalSymlinks` (on error use `filepath.Clean`); take `filepath.Rel(root, cwd)`; if the first path element matches `^level-(\d\d)$` with 1 ≤ NN ≤ 12, return NN; else `p.Current`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/game/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/game
git commit -m "feat(game): save progress and resolve the current level"
```

---

### Task 5: Level framework, build and check

**Files:**
- Create: `internal/levels/level.go`, `internal/game/build.go`
- Test: `internal/game/build_test.go`, `internal/game/play_test.go`

**Interfaces:**
- Consumes: `gitx.*` (Task 2); `EnsureRoot`, `LevelDir`, `WriteMarker`, `ReadMarker`, `SafeRemove` (Task 3).
- Produces:
  ```go
  package levels
  type StepKind int
  const ( Write StepKind = iota; Commit; Branch; Switch; ResetHard; Save )
  type Step struct { Kind StepKind; Path, Content, Msg, Name, Ref, Key string }
  func W(path, content string) Step; func C(msg string) Step; func B(name string) Step
  func S(name string) Step; func RH(ref string) Step; func Sv(key, ref string) Step
  type Level struct { N int; Title, Story, Goal string; Hints [3]string; Setup []Step; Check func(c *Ctx) (bool, string) }
  func ApplySteps(dir string, steps []Step) (refs map[string]string, err error) // SPEC §6.2 clock and table
  type Ctx struct { Dir string; Refs map[string]string }
  // methods exactly as SPEC §6.3: Head, Rev, Has, Blob, Parents, Message, Tree, Count, IsAncestor, Ignored, Merging, File, Ref
  func (c *Ctx) OnMaster() bool  // Head() == "refs/heads/master"
  func (c *Ctx) PrimaryBranch() string // keeps marked legacy levels on main checkable
  func All() []Level           // sorted by N; levels register via a package-level slice
  func ByN(n int) (Level, bool)

  package game
  func Build(root string, l levels.Level) (dir string, err error)          // SPEC §6.2 build order
  func Check(root string, l levels.Level) (ok bool, reason string, err error) // err only for missing dir/marker (RefuseError or "level N is not set up")
  ```
  Test harness in `play_test.go` (extended by Tasks 6–10):
  ```go
  type play struct { n int; fresh, solve, wrong, wrongReason string }
  var plays []play // each level task appends its rows in an init() in its own _test.go file
  func playerEnv(t *testing.T) []string // HOME, XDG_CONFIG_HOME temp; GIT_CONFIG_GLOBAL=<empty temp file>; GIT_CONFIG_NOSYSTEM=1; GIT_EDITOR=true; PATH from os
  func sh(t *testing.T, dir, script string)  // exec "sh", "-c", script in dir with playerEnv; exit code ignored
  func TestLevels(t *testing.T)               // per row, subtests "NN/fresh", "NN/solve", "NN/wrong"
  func TestLevelsDeterministic(t *testing.T)  // per row: build in two roots, compare marker Refs and `rev-parse HEAD`
  ```

- [ ] **Step 1: Write the failing tests**

`build_test.go` uses a test-only level:

```go
var demo = levels.Level{N: 9, Title: "Demo", Setup: []levels.Step{
    levels.W("a.txt", "a\n"), levels.C("First"), levels.B("side"), levels.S("side"),
    levels.W("dir/b.txt", "b\n"), levels.C("Second"), levels.Sv("side", "HEAD"), levels.S("master"), levels.RH("master"),
}, Check: func(c *levels.Ctx) (bool, string) {
    if c.Rev("side") != c.Ref("side") { return false, "side moved" }
    return c.Has("side", "dir/b.txt") && !c.Has("master", "dir/b.txt") && c.OnMaster(), "files"
}}

func TestBuildDeterministicAndIsolated(t *testing.T) {
    t.Setenv("HOME", t.TempDir())
    r1, r2 := filepath.Join(t.TempDir(), "gq"), filepath.Join(t.TempDir(), "gq")
    d1, err := Build(r1, demo); if err != nil { t.Fatal(err) }
    d2, _ := Build(r2, demo)
    m1, _ := ReadMarker(d1); m2, _ := ReadMarker(d2)
    if m1.Level != 9 || m1.Refs["side"] == "" || m1.Refs["side"] != m2.Refs["side"] { t.Fatal(m1, m2) }
    g := gitx.Git{Dir: d1, Env: gitx.CheckEnv()}
    if msg, _ := g.Run("cat-file", "commit", m1.Refs["side"]); !strings.Contains(msg, "author Gitquest Narrator <narrator@gitquest.invalid> 1767258060 +0000") { t.Fatal(msg) }
    if out, _ := g.Run("config", "--local", "user.name"); out != "Adventurer\n" { t.Fatal(out) }
    if ok, _ := g.Ok("ls-files", "--error-unmatch", ".gitquest-level"); ok { t.Fatal("marker was committed") }
    if ok, reason, err := Check(r1, demo); !ok || err != nil { t.Fatal(reason, err) }
    if _, err := Build(r1, demo); err == nil { t.Fatal("must not build over an existing level") }
}
func TestPlayerGlobalConfigIgnored(t *testing.T) {
    t.Setenv("HOME", t.TempDir())
    hostile := filepath.Join(t.TempDir(), "gitconfig")
    os.WriteFile(hostile, []byte("[commit]\n\tgpgsign = true\n[merge]\n\tff = false\n[core]\n\texcludesFile = "+filepath.Join(t.TempDir(), "ign")+"\n"), 0o644)
    t.Setenv("GIT_CONFIG_GLOBAL", hostile) // affects only the test process; Build must strip it
    root := filepath.Join(t.TempDir(), "gq")
    if _, err := Build(root, demo); err != nil { t.Fatal(err) }
    if ok, reason, _ := Check(root, demo); !ok { t.Fatal(reason) }
}
func TestCheckMissingLevel(t *testing.T) {
    root := filepath.Join(t.TempDir(), "gq"); EnsureRoot(root)
    if _, _, err := Check(root, demo); err == nil || err.Error() != "level 9 is not set up; run gitquest level 9" { t.Fatal(err) }
}
```

`play_test.go` defines the harness above with `plays` empty for now (`TestLevels` with zero rows passes trivially; it gains rows in Tasks 6–10). For each row: `fresh` builds the level and asserts `Check` returns `(false, row.fresh)`; `solve` builds a new copy, runs `sh(t, dir, row.solve)` and asserts `(true, …)`; `wrong` (if set) builds a new copy, runs `row.wrong`, and asserts `(false, row.wrongReason)`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/game/`
Expected: FAIL (build error: `undefined: levels.W` / `undefined: Build`)

- [ ] **Step 3: Implement**

`ApplySteps` keeps `clock := gitx.SetupEpoch`; every git call uses `gitx.SetupEnv(clock)`; after each `Commit`, `clock += 60`. `Build` follows the SPEC §6.2 build order exactly (marker first, `info/exclude` before any step, player config from SPEC §6.1, `core.hooksPath` = `<dir>/.git/gitquest-no-hooks`). `Check` reads the marker (missing dir → `fmt.Errorf("level %d is not set up; run gitquest level %d", n, n)`; bad marker → `RefuseError`), builds `Ctx{Dir, Refs}` and calls `l.Check`. Ctx git calls use `gitx.Git{Dir, gitx.CheckEnv()}` and prefix every command with `-c core.excludesFile=/dev/null`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/game/ ./internal/levels/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/levels internal/game
git commit -m "feat(levels): add data-driven level setup, plumbing checks and deterministic build"
```

---

### Tasks 6–10: shared shape for the level tasks (read before Task 6)

All five tasks follow the same shape. For each level `N` in the task:

**Files:**
- Create: `internal/levels/lNN_<slug>.go` (one `Level` value appended to the registry, texts verbatim from SPEC §8), `internal/game/lNN_test.go` (an `init()` appending the `play` row below)

**Interfaces:**
- Consumes: `levels.W/C/B/S/RH/Sv`, `levels.Ctx` methods, `OnMain` (Task 5); the `play` harness (Task 5).
- Produces: `levels.ByN(N)` returns the level.

Steps per task:

- [ ] **Step 1: Write the failing tests** — add the `play` rows listed for the task (exact strings below).
- [ ] **Step 2: Run them to verify they fail** — Run: `go test ./internal/game/ -run 'TestLevels'` — Expected: FAIL (`level N not registered` from a `ByN` lookup in the harness, which calls `t.Fatalf("level %d not registered", n)`).
- [ ] **Step 3: Implement each level** exactly as SPEC §8: setup steps in order, check conditions in order with their exact reasons (the "on master" condition uses `c.OnMaster()` with reason `switch back to master first`), hints verbatim.
- [ ] **Step 4: Run tests to verify they pass** — Run: `go test ./internal/game/ -run 'TestLevels'` — Expected: PASS for fresh, solve, wrong and deterministic subtests.
- [ ] **Step 5: Commit** with the message given for the task.

### Task 6: Levels 1–3

```go
{n: 1, fresh: "there are no commits yet", solve: `git add logbook.txt && git commit -q -m "First entry"`},
{n: 2, fresh: "there is no branch named explore", solve: `git switch -q -c explore`},
{n: 3, fresh: "master does not point at the same commit as repairs yet", solve: `git merge -q --ff-only repairs`,
    wrong: `git merge -q --no-ff --no-edit repairs`, wrongReason: "master got a merge commit; this merge can be a fast-forward"},
```
Extra test in `l02_test.go`: after `git branch explore` only (no switch), the reason is `the explore branch exists, but you are not on it`.
Commit: `git commit -m "feat(levels): add first commit, branch and fast-forward levels"`

### Task 7: Levels 4–5

```go
{n: 4, fresh: "master is not a merge of both captains' orders yet",
    solve: `git merge captain-b; printf 'Course: west, then east\n' > orders.txt && git add orders.txt && git commit -q --no-edit`},
{n: 5, fresh: `the last commit still says "Send telgram"`, solve: `git commit -q --amend -m "Send telegram"`,
    wrong: `git commit -q --allow-empty -m "Send telegram"`, wrongReason: `the fixed commit should sit right on top of "Add the header"; amend instead of adding a commit`},
```
Extra tests in `l04_test.go`: after only `git merge captain-b` the reason is `the merge is still in progress: fix orders.txt, git add it, then git commit`; after resolving with `printf 'Course: west\n'` the reason is `orders.txt should mention both west and east`; after committing the file with markers still in it (`git add orders.txt && git commit -q --no-edit` right after the failed merge) the reason is `orders.txt still has conflict markers in it`.
Commit: `git commit -m "feat(levels): add merge conflict and amend levels"`

### Task 8: Levels 6–7

```go
{n: 6, fresh: "nothing has been undone yet", solve: `git revert --no-edit HEAD~1`,
    wrong: `git reset -q --hard HEAD~2`, wrongReason: "history was rewritten, but the village already has those commits; run gitquest reset and try git revert"},
{n: 7, fresh: "hotfix has no new commit yet",
    solve: `git stash -q && git switch -q hotfix && echo fixed > bridge.txt && git add bridge.txt && git commit -q -m "Fix the bridge" && git switch -q master && git stash pop -q`,
    wrong: `git stash -q && git switch -q hotfix && echo fixed > bridge.txt && git add bridge.txt && git commit -q -m "Fix the bridge" && git switch -q master`,
    wrongReason: "your plan B draft is not back in notes.txt yet"},
```
Extra test in `l07_test.go`: on the fresh level, `git switch hotfix` exits non-zero (proves the stash is needed).
Commit: `git commit -m "feat(levels): add revert and stash levels"`

### Task 9: Levels 8–9

```go
{n: 8, fresh: "there is no branch named treasure", solve: `git branch treasure HEAD@{1}`,
    wrong: `git branch treasure`, wrongReason: "treasure exists, but it does not contain the lost map commit"},
{n: 9, fresh: "master has no sourdough.txt yet", solve: `git cherry-pick experiments~1`,
    wrong: `git merge -q --no-edit experiments`, wrongReason: "master got a merge commit; copy just the one commit instead"},
```
Commit: `git commit -m "feat(levels): add reflog rescue and cherry-pick levels"`

### Task 10: Levels 10–12

```go
{n: 10, fresh: `campfire still contains the "Scout the trail" commit`, solve: `git rebase -q --onto master trail campfire`,
    wrong: `git rebase -q master`, wrongReason: "campfire should have exactly 2 commits on top of master"},
{n: 11, fresh: `there should be exactly one commit after "Write chapter 1"`,
    solve: `git reset -q --soft HEAD~2 && git commit -q -m "Write chapters 2 and 3"`,
    wrong: `git reset -q --hard HEAD~2 && git commit -q --allow-empty -m "Write chapters 2 and 3"`,
    wrongReason: "your uncommitted edit to ch1.txt is gone (git reset --hard throws working changes away); run gitquest reset and try --soft"},
{n: 12, fresh: "there is no committed .gitignore yet",
    solve: `printf '*.log\ncache/\nsecrets.env\n' > .gitignore && git add .gitignore notes.md && git commit -q -m "Ignore the junk"`,
    wrong: `printf '*.log\ncache/\n' > .gitignore && git add .gitignore notes.md && git commit -q -m "Ignore some"`,
    wrongReason: "secrets.env is not ignored by your .gitignore"},
```
Extra tests in `l11_test.go`: `git reset -q --mixed HEAD~2 && git commit -q -m "Write chapters 2 and 3"` fails with `there should be exactly one commit after "Write chapter 1"` (nothing staged, so the commit fails and HEAD stays on chapter 1). In `l12_test.go`: `printf '*.tmp\n' > .gitignore && git add -A && git commit -q -m all` fails with `debug.log was committed; remove it with git rm --cached debug.log and commit again`. In `levels_test.go` (package `levels`): `len(All()) == 12`, `All()[i].N == i+1`, and every level has non-empty Title, Story, Goal and three non-empty hints.
Commit: `git commit -m "feat(levels): add rebase onto, reset and gitignore levels"`

---

### Task 11: `start`, `level` and `status`

**Files:**
- Modify: `internal/cli/cli.go`
- Test: `internal/cli/play_test.go`

**Interfaces:**
- Consumes: `game.Root`, `game.StateDir`, `game.EnsureRoot`, `game.Build`, `game.LevelDir`, `game.ReadMarker`, `game.DisplayPath`, `game.LoadProgress`, `game.SaveProgress`, `levels.ByN`, `levels.All`, `gitx.VersionOK`.
- Produces: unexported `readyText(l levels.Level, dirShown string, existed bool) string` (SPEC §7.1 text), dispatch for `start`, `level`, `status`.

- [ ] **Step 1: Write the failing tests**

```go
func TestStartFirstLevel(t *testing.T) {
    d, _, out, _ := testDeps(t)
    if Run([]string{"start"}, d) != 0 { t.Fatal(out) }
    l, _ := levels.ByN(1)
    want := "Level 1/12 · The Empty Logbook\n\n" + l.Story + "\n\nGoal: " + l.Goal + "\n\nYour level is ready:\n  cd ~/gq/level-01\nRun gitquest check when you think you're done, or gitquest hint if you're stuck.\n"
    if out.String() != want { t.Fatalf("got\n%s\nwant\n%s", out, want) }
}
func TestLevelTwiceKeepsWork(t *testing.T) {
    d, home, out, _ := testDeps(t)
    Run([]string{"level", "3"}, d)
    work := filepath.Join(home, "gq", "level-03", "mine.txt"); os.WriteFile(work, []byte("x"), 0o644)
    out.Reset()
    if Run([]string{"level", "3"}, d) != 0 || !strings.Contains(out.String(), "Your level is already set up (gitquest reset rebuilds it):\n") { t.Fatal(out) }
    if _, err := os.Stat(work); err != nil { t.Fatal("player work was lost") }
}
func TestLevelBadNumberAndForeignDir(t *testing.T) {
    d, home, _, errb := testDeps(t)
    for _, a := range []string{"0", "13", "x"} {
        errb.Reset()
        if Run([]string{"level", a}, d) != 2 || errb.String() != "gitquest: level must be a number from 1 to 12\n" { t.Fatal(a, errb) }
    }
    root := filepath.Join(home, "gq"); stray := filepath.Join(root, "level-02")
    os.MkdirAll(stray, 0o755) // root exists, is non-empty and has no root marker
    errb.Reset()
    if Run([]string{"level", "2"}, d) != 1 || errb.String() != "gitquest: "+root+" already exists and is not a gitquest folder (no .gitquest-root); set GITQUEST_ROOT to another path\n" { t.Fatal(errb) }
    os.WriteFile(filepath.Join(root, ".gitquest-root"), []byte("gitquest game root v1\n"), 0o644) // now a real root with an unmarked level dir
    errb.Reset()
    if Run([]string{"level", "2"}, d) != 1 || errb.String() != "gitquest: refusing to touch "+stray+": it has no .gitquest-level marker; move or delete it yourself\n" { t.Fatal(errb) }
    if entries, _ := os.ReadDir(stray); len(entries) != 0 { t.Fatal("wrote into the stray dir") }
}
func TestStatus(t *testing.T) {
    d, _, out, _ := testDeps(t)
    stateDir := d.Getenv("GITQUEST_HOME")
    p, _, _ := game.LoadProgress(stateDir, time.Time{})
    p.Level(1).SolvedAt = time.Unix(1, 0); p.Level(2).SolvedAt = time.Unix(1, 0); p.Current = 3; p.Level(3).Hints = 2; p.Level(4).Hints = 1
    game.SaveProgress(stateDir, p)
    Run([]string{"status"}, d)
    lines := strings.Split(out.String(), "\n")
    want := []string{"gitquest · 2/12 solved", "", "✓  1 The Empty Logbook", "✓  2 A Fork in the Path", "▶  3 Catching Up  (2 hints used)", "   4 Two Captains  (1 hint used)", "   5 Typo in the Telegram"}
    if !reflect.DeepEqual(lines[:7], want) || len(lines) != 15 { t.Fatalf("%q", lines) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/`
Expected: FAIL (`unknown command "start"`)

- [ ] **Step 3: Implement per SPEC §7.1 and §7.4**

Order for `start`/`level`: parse N (usage error), `gitx.VersionOK` (exit 1), `game.Root` (relative → usage error 2 with `gitquest: GITQUEST_ROOT must be an absolute path`), `EnsureRoot`, existing dir → `ReadMarker` must say `gitquest` and level N (else the `RefuseError` text), else `Build`; then `Current = N`, save progress, print. Print the corrupt-progress note to stderr when `movedTo != ""`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): add start, level and status commands"
```

---

### Task 12: `check`, `hint` and `reset`

**Files:**
- Modify: `internal/cli/cli.go`
- Test: `internal/cli/play_test.go`

**Interfaces:**
- Consumes: Task 11 helpers; `game.Check`, `game.SafeRemove`, `game.CurrentLevel`.
- Produces: dispatch for `check`, `hint`, `reset`.

- [ ] **Step 1: Write the failing tests**

```go
func playerSh(t *testing.T, dir, script string) { /* sh -c in dir with HOME/GIT_CONFIG_GLOBAL temp, GIT_CONFIG_NOSYSTEM=1, GIT_EDITOR=true */ }

func TestCheckFailThenPass(t *testing.T) {
    d, home, out, errb := testDeps(t)
    if Run([]string{"check"}, d) != 1 || errb.String() != "gitquest: no level in progress; run gitquest start\n" { t.Fatal(errb) }
    Run([]string{"start"}, d); out.Reset()
    if Run([]string{"check"}, d) != 1 || out.String() != "✗ Not solved yet: there are no commits yet\nStuck? Try gitquest hint\n" { t.Fatal(out) }
    dir := filepath.Join(home, "gq", "level-01")
    playerSh(t, dir, `git add logbook.txt && git commit -q -m "First entry"`)
    out.Reset()
    if Run([]string{"check"}, d) != 0 || out.String() != "✓ Level 1 solved: The Empty Logbook\nNext: gitquest level 2\n" { t.Fatal(out) }
}
func TestCheckUsesCwdLevel(t *testing.T) {
    d, home, out, _ := testDeps(t)
    Run([]string{"level", "2"}, d); Run([]string{"level", "1"}, d) // current = 1
    dir := filepath.Join(home, "gq", "level-02")
    playerSh(t, dir, `git switch -q -c explore`)
    d.Getwd = func() (string, error) { return dir, nil }
    out.Reset()
    if Run([]string{"check"}, d) != 0 || !strings.HasPrefix(out.String(), "✓ Level 2 solved") { t.Fatal(out) }
}
func TestHints(t *testing.T) {
    d, _, out, _ := testDeps(t)
    Run([]string{"level", "8"}, d)
    l, _ := levels.ByN(8)
    for i := 0; i < 3; i++ {
        out.Reset(); Run([]string{"hint"}, d)
        if out.String() != fmt.Sprintf("Hint %d/3: %s\n", i+1, l.Hints[i]) { t.Fatal(out) }
    }
    out.Reset(); Run([]string{"hint"}, d)
    want := fmt.Sprintf("Hint 1/3: %s\nHint 2/3: %s\nHint 3/3: %s\nThat's every hint. You've got this!\n", l.Hints[0], l.Hints[1], l.Hints[2])
    if out.String() != want { t.Fatal(out) }
}
func TestResetRebuilds(t *testing.T) {
    d, home, out, _ := testDeps(t)
    Run([]string{"level", "6"}, d)
    dir := filepath.Join(home, "gq", "level-06")
    playerSh(t, dir, `git reset -q --hard HEAD~2`)
    out.Reset()
    if Run([]string{"reset"}, d) != 0 || !strings.HasPrefix(out.String(), "Level 6 rebuilt.\n\nLevel 6/12 · Undo the Spell\n") { t.Fatal(out) }
    out.Reset()
    if Run([]string{"check"}, d) != 1 || !strings.Contains(out.String(), "nothing has been undone yet") { t.Fatal(out) }
}
func TestResetRefusesUnmarked(t *testing.T) {
    d, home, _, errb := testDeps(t)
    Run([]string{"level", "4"}, d)
    dir := filepath.Join(home, "gq", "level-04")
    os.Remove(filepath.Join(dir, ".gitquest-level"))
    if Run([]string{"reset", "4"}, d) != 1 || errb.String() != "gitquest: refusing to touch "+dir+": it has no .gitquest-level marker; move or delete it yourself\n" { t.Fatal(errb) }
    if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil { t.Fatal("deleted an unmarked dir") }
}
func TestLastLevelMessage(t *testing.T) {
    d, home, out, _ := testDeps(t)
    Run([]string{"level", "12"}, d)
    playerSh(t, filepath.Join(home, "gq", "level-12"), `printf '*.log\ncache/\nsecrets.env\n' > .gitignore && git add .gitignore notes.md && git commit -q -m x`)
    out.Reset(); Run([]string{"check"}, d)
    if out.String() != "✓ Level 12 solved: Secrets in the Attic\nThat was the last level. You escaped gitquest!\n" { t.Fatal(out) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/`
Expected: FAIL (`unknown command "check"`)

- [ ] **Step 3: Implement per SPEC §7.2, §7.3, §7.5**

Current level via `game.CurrentLevel(root, cwd, progress)` (Getwd error → use `progress.Current`). `reset [N]` parses N like `level`. Paths in error messages are the absolute dir (not `~`).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): add check, hint and reset commands"
```

---

### Task 13: README and final verification

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README.md** with: first line `gitquest: learn git by escaping puzzle rooms`; Install (`go install github.com/dpologdvinity/gitquest@latest`, release archives; needs git ≥ 2.32); How to play (start → cd → real git → check → hint), a real transcript of level 1 copied from your own run; the 12 level titles with the git concept each teaches; Safety (game root, `.gitquest-root`/`.gitquest-level` markers, gitquest never touches other folders, level repos have their own identity `Adventurer` and no hooks); Environment (`GITQUEST_ROOT`, `GITQUEST_HOME`); License MIT.

- [ ] **Step 2: Full verification**

Run: `test -z "$(gofmt -l .)" && go vet ./... && go test ./... && go build -o /tmp/gitquest .`
Expected: all pass.

- [ ] **Step 3: Manual playthrough (SPEC §11)**

With `export GITQUEST_ROOT=$(mktemp -d)/gq GITQUEST_HOME=$(mktemp -d)`: `/tmp/gitquest level 1`, solve it, `check`; same for level 4 (resolve the conflict in an editor) and level 8 (use `git reflog`). Expected: each prints `✓ Level N solved: …`; `/tmp/gitquest status` shows 3/12.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add README with how to play and safety notes"
```
