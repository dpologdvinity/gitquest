# gitquest — Specification (v0.1)

`gitquest` is an escape-room puzzle game that teaches git. Each level is a real
git repository with a small story and a broken or unfinished state. You solve
it with real `git` commands in your own shell, then run `gitquest check`.
Checks read the repository with git plumbing only, never by parsing porcelain
output such as `git status` or `git log`.

This document is the source of truth for v0.1 behavior. `docs/PLAN.md` says how
to build it. If the plan and the spec disagree, the spec wins.

## 1. Goals and non-goals

### Goals (v0.1)
- 12 levels (§8): first commit, branch, fast-forward merge, merge conflict, amend, revert, stash, reflog rescue, cherry-pick, `rebase --onto`, `reset --soft` vs `--hard`, `.gitignore`.
- Commands: `start`, `level N`, `check`, `hint` (3 progressive hints), `status`, `reset [N]`.
- Deterministic level setup: fixed identity, fixed dates, isolated git config. The same level always produces the same commit IDs.
- Safety: gitquest creates and deletes only inside its own game root, and refuses any directory without its marker file (§5).
- Every level has a test that applies the intended solution and passes, and a test that the unsolved state fails.

### Non-goals (v0.1)
- No network, no remotes, no `push`/`pull`/`fetch` levels.
- No interactive rebase level, no submodules, no hooks.
- No TUI, no timer, no scoring beyond solved/hints used.
- No Windows build (git config isolation uses `/dev/null`; see §11 Later).

## 2. Name availability (checked 2026-10-23)

| Place | Result |
|---|---|
| GitHub `dpologdvinity/gitquest` | free (404) |
| Homebrew formula / cask | free (404 / 404) |
| npm `gitquest` | free (404) |
| GitHub repos named like it | about 877 results; the largest is `Kashan-2912/gitquest` (36 stars, a GitHub-profile RPG, unrelated). The name is common but not owned by a known tool. **Risk: low-medium**. A search for "gitquest" will be crowded, so the README should say "gitquest: learn git by escaping puzzle rooms" near the top. |

## 3. Tech stack (fixed)

| Item | Choice |
|---|---|
| Language | Go. `go.mod` says `go 1.26` (builds with the installed 1.27.3). |
| Module path | `github.com/dpologdvinity/gitquest` |
| Dependencies | none. Stdlib only. The game shells out to the user's `git` (≥ 2.32, needed for `GIT_CONFIG_GLOBAL`). |
| Tests | stdlib `testing`, real `git` in temp dirs |
| Release | GoReleaser v2, linux/darwin × amd64/arm64, `CGO_ENABLED=0` |
| CI | GitHub Actions: `actions/checkout@v7`, `actions/setup-go@v7`, `goreleaser/goreleaser-action@v7` with `version: "~> v2"` (latest tags on 2026-10-23: v7.0.1, v7.0.0, v7.3.0; GoReleaser v2.19.0) |
| License | MIT, `Copyright (c) 2026 Kaitlyn Bassford` |

No third-party module is needed. Using a Go git library (go-git) was rejected:
the game must check exactly what the real `git` the player uses did, and
go-git would add a large dependency.

## 4. CLI interface

```
gitquest start           set up the first unsolved level
gitquest level <N>       set up level N (1-12)
gitquest check           check the current level
gitquest hint            show the next hint for the current level
gitquest status          show progress for all levels
gitquest reset [N]       delete and rebuild level N (default: current level)
gitquest --version | -V  print "gitquest <version>"
gitquest --help | -h | help
```

### 4.1 Usage text (exact)

```
gitquest - learn git by escaping puzzle rooms

usage:
  gitquest start          set up the first unsolved level
  gitquest level <N>      set up level N (1-12)
  gitquest check          check your solution
  gitquest hint           get a hint (3 per level)
  gitquest status         show your progress
  gitquest reset [N]      rebuild a level from scratch
  gitquest --version      print the version

Levels live in ~/gitquest (set GITQUEST_ROOT to change it).
Solve each level with real git commands inside its folder.
```

### 4.2 Exit codes and errors

| Code | When |
|---|---|
| 0 | Success, including `check` that passes. |
| 1 | `check` did not pass; no level in progress; level not set up; safety refusal; git missing or too old; I/O errors. |
| 2 | Usage error. |

Errors print `gitquest: <message>` to stderr. Exact messages:
- unknown command: `gitquest: unknown command "<x>" (see gitquest --help)`
- bad level number: `gitquest: level must be a number from 1 to 12`
- git missing/old: `gitquest: needs git 2.32 or newer on your PATH`
- no current level: `gitquest: no level in progress; run gitquest start`
- level not built (check): `gitquest: level <N> is not set up; run gitquest level <N>`
- root refusal: `gitquest: <root> already exists and is not a gitquest folder (no .gitquest-root); set GITQUEST_ROOT to another path`
- level dir refusal: `gitquest: refusing to touch <dir>: <reason>; move or delete it yourself` where reason is one of `it has no .gitquest-level marker`, `it is a symlink`, `it is outside the game root`, `its marker is for level <M>`.

## 5. Files and safety

### 5.1 Locations
- Game root: `$GITQUEST_ROOT` if set and non-empty (must be absolute, else usage error `gitquest: GITQUEST_ROOT must be an absolute path`), else `$HOME/gitquest`.
- Level dir: `<root>/level-NN` (`NN` two digits, `level-01` … `level-12`).
- State dir: `$GITQUEST_HOME` if set, else `$XDG_STATE_HOME/gitquest` if absolute, else `$HOME/.local/state/gitquest`. Contains `progress.json` (dir `0700`, file `0600`, atomic write via temp file + rename).
- Paths shown to the user replace a leading `$HOME/` with `~/`.

### 5.2 Markers
- Root marker: `<root>/.gitquest-root`, content exactly `gitquest game root v1\n`.
- Level marker: `<level dir>/.gitquest-level`, JSON `{"game":"gitquest","level":N,"refs":{...}}`. `refs` maps names to commit IDs saved during setup (§6.2). It is written **before** `git init` (with empty refs) and rewritten at the end of setup. Setup appends `.gitquest-level` to `.git/info/exclude` so the player never sees it in `git status`.

### 5.3 Rules
1. `EnsureRoot(root)`: if `root` does not exist, create it (`0755`) and write the root marker. If it exists and is a directory with the exact root marker, use it. If it exists, is a directory, and is empty, write the marker. Anything else (a file, or a non-empty dir without the marker) is the root refusal error.
2. Gitquest only creates `<root>`, `<root>/.gitquest-root`, and `<root>/level-NN` trees.
3. `SafeRemove(root, n)` deletes `<root>/level-NN` only if **all** hold, else returns the level dir refusal:
   - the root has the exact root marker;
   - `os.Lstat(dir)` is a directory and not a symlink;
   - `filepath.EvalSymlinks(filepath.Dir(dir)) == filepath.EvalSymlinks(root)`;
   - `.gitquest-level` exists, parses, has `"game":"gitquest"` and `"level":n`.
   Then `os.RemoveAll(dir)`.
4. Building a level whose dir already exists never overwrites it (`start`/`level` print the "already set up" text; `reset` uses `SafeRemove` first).
5. If setup fails halfway, the half-built dir is removed with `SafeRemove` (the marker was written first, so this is allowed).

### 5.4 `progress.json`

```go
type Progress struct {
    Version int                    `json:"version"` // 1
    Current int                    `json:"current"` // 0 = none
    Levels  map[string]*LevelState `json:"levels"`  // "1".."12"
}
type LevelState struct {
    SolvedAt time.Time `json:"solved_at"` // zero if unsolved; first solve time is kept
    Hints    int       `json:"hints"`     // 0..3
}
```
Missing file: empty progress. Unparseable: rename to `progress.json.corrupt-<unix>` and start empty, with the stderr note `gitquest: progress file was unreadable; moved it to <path> and started fresh`.

### 5.5 Current level
1. If the working directory (after `EvalSymlinks`) is `<root>/level-NN` or inside it (root also resolved), the current level is `NN`.
2. Otherwise `progress.Current`. If that is 0: `no current level` error.

## 6. Git usage (package `internal/gitx`)

### 6.1 Environments
- **Setup env** (building levels): `os.Environ()` with every `GIT_*` variable removed, plus `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_AUTHOR_NAME=Gitquest Narrator`, `GIT_AUTHOR_EMAIL=narrator@gitquest.invalid`, `GIT_COMMITTER_NAME=Gitquest Narrator`, `GIT_COMMITTER_EMAIL=narrator@gitquest.invalid`, `GIT_AUTHOR_DATE` and `GIT_COMMITTER_DATE` = `@<unix> +0000` (§6.2), `LC_ALL=C`, `TZ=UTC`, `GIT_TERMINAL_PROMPT=0`.
- **Check env**: `os.Environ()` minus `GIT_*`, plus `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=/dev/null`, `LC_ALL=C`, `GIT_OPTIONAL_LOCKS=0`. Check commands also pass `-c core.excludesFile=/dev/null` so the player's personal ignore file does not change results.
- **Player config** written into every level repo by setup with `git config` (so the player's own global config cannot break a level): `user.name Adventurer`, `user.email adventurer@gitquest.invalid`, `commit.gpgsign false`, `tag.gpgsign false`, `merge.ff true`, `core.autocrlf false`, `core.hooksPath <level dir>/.git/gitquest-no-hooks` (a path that does not exist, so no hooks run).

### 6.2 Deterministic setup
- `git init -q -b main` in the level dir.
- A setup clock starts at `1767258000` (2026-01-01 09:00:00 UTC). Every setup git command runs with both dates set to the clock; each `Commit` step advances the clock by 60 s **after** committing.
- Setup steps (data, applied in order):

| Step | Effect |
|---|---|
| `Write{Path, Content}` | create parent dirs, write file `0644` |
| `Commit{Msg}` | `git add -A`, `git commit -q -m <Msg>` |
| `Branch{Name}` | `git branch <Name>` |
| `Switch{Name}` | `git switch -q <Name>` |
| `ResetHard{Ref}` | `git reset -q --hard <Ref>` |
| `Save{Key, Ref}` | `refs[Key] = git rev-parse <Ref>` |

Build order (`game.Build`), exactly:
1. `EnsureRoot`; fail if the level dir exists.
2. `os.Mkdir(dir, 0755)`; write `.gitquest-level` with empty `refs`.
3. `git init -q -b main`; append the line `.gitquest-level` to `.git/info/exclude` (so `git add -A` never picks it up).
4. Write the player config (§6.1) with `git config`.
5. Apply the level's steps (§6.2 table).
6. Rewrite `.gitquest-level` with the saved `refs`.
On any error after step 2: `SafeRemove`, then return the error.

Building the same level twice gives identical commit IDs (tested).

### 6.3 Plumbing used by checks

`Ctx` helper methods (all with the check env; a failing git command yields the zero value):

| Method | Git |
|---|---|
| `Head() string` | `symbolic-ref -q HEAD` (e.g. `refs/heads/main`; `""` if detached) |
| `Rev(name) string` | `rev-parse -q --verify <name>^{commit}` (`""` if missing) |
| `Has(rev, path) bool` | `rev-parse -q --verify <rev>:<path>` succeeds |
| `Blob(rev, path) string` | `cat-file blob <rev>:<path>` |
| `Parents(rev) []string` | `rev-list --parents -n 1 <rev>`, fields after the first |
| `Message(rev) string` | `cat-file commit <rev>`, text after the first empty line, `TrimSpace` |
| `Tree(rev) string` | `rev-parse <rev>^{tree}` |
| `Count(args...) int` | `rev-list --count <args>` |
| `IsAncestor(a, b) bool` | `merge-base --is-ancestor <a> <b>` exit 0 |
| `Ignored(path) bool` | `check-ignore -q <path>` exit 0 |
| `Merging() bool` | `rev-parse -q --verify MERGE_HEAD` succeeds |
| `File(path) (string, bool)` | `os.ReadFile(<level dir>/<path>)` (working tree, not git) |
| `Ref(key) string` | `marker.refs[key]` |

Version check: `git version` output must match `^git version (\d+)\.(\d+)` with (major, minor) ≥ (2, 32).

## 7. Commands

### 7.1 `start` and `level N`
`start` picks the lowest level without `SolvedAt`; if all 12 are solved it prints
`You have solved all 12 levels. Replay any of them with gitquest level <N>.` and exits 0.
Both then: check git version, `EnsureRoot`, build the level unless its dir
exists (a dir without a valid marker is the refusal error), set
`progress.Current = N`, save, and print (exact):

```
Level 3/12 · Catching Up

<Story>

Goal: <Goal>

Your level is ready:
  cd ~/gitquest/level-03
Run gitquest check when you think you're done, or gitquest hint if you're stuck.
```

If the dir already existed, the line `Your level is ready:` becomes
`Your level is already set up (gitquest reset rebuilds it):`.

### 7.2 `check`
Resolve the current level (§5.5). The level dir must exist with a valid marker.
Run the level's `Check`. On pass print:

```
✓ Level 3 solved: Catching Up
Next: gitquest level 4
```
(for level 12 the second line is `That was the last level. You escaped gitquest!`),
set `SolvedAt` if zero, set `Current` to N, exit 0. On fail print
`✗ Not solved yet: <reason>` and `Stuck? Try gitquest hint` on stdout, exit 1.

### 7.3 `hint`
Resolve the current level. With `h = Hints`: if `h < 3`, set `h+1`, save, print
`Hint <h+1>/3: <text>`. If `h == 3`, print all three lines `Hint k/3: <text>`
and then `That's every hint. You've got this!`.

### 7.4 `status`
```
gitquest · 2/12 solved

✓  1 The Empty Logbook
✓  2 A Fork in the Path
▶  3 Catching Up  (2 hints used)
   4 Two Captains
…
```
Each line: `fmt.Sprintf("%s %2d %s", mark, n, title)` plus `  (1 hint used)` or
`  (<k> hints used)` when `k > 0`. Mark: `✓` solved, `▶` current and unsolved,
otherwise a space.

### 7.5 `reset [N]`
`N` defaults to the current level. If the dir exists, `SafeRemove` it, then build
again, set `Current = N`, and print `Level N rebuilt.` followed by an empty line
and the §7.1 text ("ready" form). Hints and `SolvedAt` are kept.

## 8. Levels (package `internal/levels`)

```go
type Level struct {
    N           int
    Title, Story, Goal string
    Hints       [3]string
    Setup       []Step
    Check       func(c *Ctx) (ok bool, reason string)
}
```

Every check first requires `c.Head() == "refs/heads/main"` (reason
`switch back to main first`) unless stated. Conditions are tested in the order
listed; the first failing one gives the reason. Text in this section is exact.

### Level 1 · The Empty Logbook
- Story: `You are the new keeper of the gitquest lighthouse. The old keeper left a logbook on the desk, but nothing in it was ever saved.`
- Goal: `Save logbook.txt in a commit on the main branch.`
- Setup: `Write logbook.txt "Day 1: the lamp is lit.\n"`.
- Check: (1) `Rev("HEAD") != ""` else `there are no commits yet`; (2) on main; (3) `Has("HEAD","logbook.txt")` else `the latest commit does not contain logbook.txt`.
- Hints: `A commit is a saved snapshot. First tell git which files belong in it.` / `git add logbook.txt stages the file for the next commit.` / `git add logbook.txt && git commit -m "First entry"`
- Solution: `git add logbook.txt`; `git commit -m "First entry"`.

### Level 2 · A Fork in the Path
- Story: `Behind the lighthouse the path splits in two. You want to explore the new trail without scribbling on the main map.`
- Goal: `Create a branch named explore and switch to it.`
- Setup: `Write map.txt "lighthouse\n"`, `Commit "Draw the map"`.
- Check (no main requirement): (1) `Rev("refs/heads/explore") != ""` else `there is no branch named explore`; (2) `Head() == "refs/heads/explore"` else `the explore branch exists, but you are not on it`.
- Hints: `Branches are movable name tags for commits.` / `git branch explore creates it; git switch explore moves you onto it.` / `git switch -c explore`
- Solution: `git switch -c explore`.

### Level 3 · Catching Up
- Story: `Your crew patched the raft on the repairs branch. The main branch still points at the old, leaky raft.`
- Goal: `Bring main up to date with repairs without creating a merge commit.`
- Setup: `Write raft.txt "logs\n"`, `Commit "Build the raft"`, `Branch repairs`, `Switch repairs`, `Write sail.txt "patched\n"`, `Commit "Patch the sail"`, `Write knots.txt "tight\n"`, `Commit "Tie the knots"`, `Save repairs repairs`, `Switch main`.
- Check: (1) on main; (2) `len(Parents("main")) < 2` else `main got a merge commit; this merge can be a fast-forward`; (3) `Rev("main") == Ref("repairs")` else `main does not point at the same commit as repairs yet`.
- Hints: `When main has no new commits of its own, a merge can just slide main forward.` / `Stay on main and merge the repairs branch.` / `git merge --ff-only repairs`
- Solution: `git merge --ff-only repairs`. Wrong solution test: `git merge --no-ff --no-edit repairs` fails with reason (2).

### Level 4 · Two Captains
- Story: `Two captains wrote different courses into orders.txt. The ship can only follow one set of orders, so you must merge them.`
- Goal: `Merge captain-b into main. Resolve the conflict so orders.txt mentions both west and east, then finish the merge.`
- Setup: `Write orders.txt "Course: north\n"`, `Commit "Agree on a course"`, `Branch captain-b`, `Write orders.txt "Course: west\n"`, `Commit "Captain A: sail west"`, `Save main main`, `Switch captain-b`, `Write orders.txt "Course: east\n"`, `Commit "Captain B: sail east"`, `Save b captain-b`, `Switch main`.
- Check: (1) on main; (2) `!Merging()` else `the merge is still in progress: fix orders.txt, git add it, then git commit`; (3) `Parents("HEAD") == [Ref("main"), Ref("b")]` else `main is not a merge of both captains' orders yet`; (4) no line of `Blob("HEAD","orders.txt")` starts with `<<<<<<<`, `=======` or `>>>>>>>` else `orders.txt still has conflict markers in it`; (5) the blob, lowercased, contains `west` and `east` else `orders.txt should mention both west and east`.
- Hints: `Start with git merge captain-b. Git stops and marks the conflict inside orders.txt.` / `Edit orders.txt: delete the <<<<<<<, ======= and >>>>>>> lines and write a course that has both west and east.` / `git merge captain-b, edit orders.txt, then git add orders.txt && git commit --no-edit`
- Solution: `git merge captain-b` (exits 1), write `Course: west, then east\n`, `git add orders.txt`, `git commit --no-edit`.

### Level 5 · Typo in the Telegram
- Story: `You just sent a telegram to HQ, but the commit message has a typo. Nobody has seen it yet, so you can fix it in place.`
- Goal: `Change the last commit's message to exactly "Send telegram" without adding a new commit.`
- Setup: `Write header.txt "TO: HQ\n"`, `Commit "Add the header"`, `Save base HEAD`, `Write telegram.txt "ARRIVING TUESDAY STOP\n"`, `Commit "Send telgram"`, `Save typo HEAD`.
- Check: (1) on main; (2) `Rev("HEAD") != Ref("typo")` else `the last commit still says "Send telgram"`; (3) `Parents("HEAD") == [Ref("base")]` else `the fixed commit should sit right on top of "Add the header"; amend instead of adding a commit`; (4) `Message("HEAD") == "Send telegram"` else `the message is "<msg>", not "Send telegram"`; (5) `Tree("HEAD") == Tree(Ref("typo"))` else `the files changed too; only the message should change`.
- Hints: `You can rewrite the most recent commit as long as nobody else has it yet.` / `git commit --amend replaces the last commit with a new one.` / `git commit --amend -m "Send telegram"`
- Solution: `git commit --amend -m "Send telegram"`. Wrong solution test: `git commit --allow-empty -m "Send telegram"` fails with reason (3).

### Level 6 · Undo the Spell
- Story: `Someone cast a growth spell on the village garden, and everyone already pulled it. You must undo it without rewriting shared history.`
- Goal: `Undo the "Cast a growth spell" commit with a new commit, keeping every existing commit.`
- Setup: `Write garden.txt "roses\n"`, `Commit "Plant the garden"`, `Write spell.txt "grow grow grow\n"`, `Commit "Cast a growth spell"`, `Write garden.txt "roses\nwatered\n"`, `Commit "Water the plants"`, `Save tip HEAD`.
- Check: (1) on main; (2) `IsAncestor(Ref("tip"), "HEAD")` else `history was rewritten, but the village already has those commits; run gitquest reset and try git revert`; (3) `Rev("HEAD") != Ref("tip")` else `nothing has been undone yet`; (4) `!Has("HEAD","spell.txt")` else `spell.txt is still in the latest commit`; (5) `Blob("HEAD","garden.txt") == "roses\nwatered\n"` else `garden.txt lost its watering; undo only the spell`.
- Hints: `Shared history should not be rewritten. Add a commit that does the opposite instead.` / `git log --oneline shows the commit to undo; it is HEAD~1.` / `git revert --no-edit HEAD~1`
- Solution: `git revert --no-edit HEAD~1`. Wrong solution test: `git reset --hard HEAD~2` fails with reason (2).

### Level 7 · Pocket Dimension
- Story: `You were drafting plan B when a messenger burst in: the bridge is broken! Your draft is not ready to commit, and git will not let you switch to the hotfix branch with it lying around.`
- Goal: `Put your draft aside, commit a file named bridge.txt on the hotfix branch, then return to main and bring your draft back (uncommitted).`
- Setup: `Write notes.txt "plan A\n"`, `Commit "Write plan A"`, `Branch hotfix`, `Switch hotfix`, `Write notes.txt "plan A\nbridge: broken\n"`, `Commit "Report the broken bridge"`, `Save hotfix hotfix`, `Switch main`, `Save main main`, `Write notes.txt "plan A\nplan B (draft)\n"`.
- Check (main requirement is condition 4): (1) `Rev("hotfix") != Ref("hotfix")` else `hotfix has no new commit yet`; (2) `IsAncestor(Ref("hotfix"), "hotfix")` else `the hotfix branch lost its earlier commit`; (3) `Has("hotfix","bridge.txt")` else `the newest hotfix commit has no bridge.txt`; (4) on main; (5) `Rev("main") == Ref("main")` else `main changed; your draft should stay uncommitted`; (6) `File("notes.txt") == "plan A\nplan B (draft)\n"` else `your plan B draft is not back in notes.txt yet`.
- Hints: `git stash tucks uncommitted changes away and leaves a clean working tree.` / `Stash, switch to hotfix, add and commit bridge.txt, switch back to main, then bring the stash back.` / `git stash && git switch hotfix && echo fixed > bridge.txt && git add bridge.txt && git commit -m "Fix the bridge" && git switch main && git stash pop`
- Solution: exactly hint 3.

### Level 8 · The Lost Treasure Map
- Story: `Last night you drew a treasure map, then panicked and ran git reset --hard. The commit looks gone, but git quietly remembers everywhere HEAD has been.`
- Goal: `Find the lost commit and create a branch named treasure that points at it.`
- Setup: `Write journal.txt "set sail\n"`, `Commit "Start the journal"`, `Write treasure-map.txt "X marks the spot\n"`, `Commit "Draw the treasure map"`, `Save lost HEAD`, `ResetHard HEAD~1`.
- Check (no main requirement): (1) `Rev("refs/heads/treasure") != ""` else `there is no branch named treasure`; (2) `IsAncestor(Ref("lost"), "refs/heads/treasure")` else `treasure exists, but it does not contain the lost map commit`.
- Hints: `git reflog lists every commit HEAD has pointed at, even after a reset.` / `Find "Draw the treasure map" in git reflog. HEAD@{1} is where HEAD was before the reset.` / `git branch treasure HEAD@{1}`
- Solution: `git branch treasure HEAD@{1}`.

### Level 9 · Borrowed Recipe
- Story: `The experiments branch has three new recipes. Only the sourdough is good enough for the bakery menu.`
- Goal: `Copy only the "Perfect the sourdough" commit onto main.`
- Setup: `Write menu.txt "bread\n"`, `Commit "Open the bakery"`, `Branch experiments`, `Switch experiments`, `Write glitter.txt "glitter frosting\n"`, `Commit "Try glitter frosting"`, `Write sourdough.txt "flour, water, salt, patience\n"`, `Commit "Perfect the sourdough"`, `Save sourdough HEAD`, `Write fishcake.txt "fish cake\n"`, `Commit "Try fish cake"`, `Switch main`, `Write sign.txt "OPEN\n"`, `Commit "Paint the sign"`, `Save main HEAD`.
- Check: (1) on main; (2) `IsAncestor(Ref("main"), "main")` else `main lost its own commits`; (3) `Count("--min-parents=2", Ref("main")+"..main") == 0` else `main got a merge commit; copy just the one commit instead`; (4) `Has("main","sourdough.txt")` and its blob equals `Blob(Ref("sourdough"),"sourdough.txt")` else `main has no sourdough.txt yet`; (5) neither `glitter.txt` nor `fishcake.txt` in `main` else `main picked up recipes it should not have`.
- Hints: `git cherry-pick copies a single commit onto the branch you are on.` / `git log --oneline experiments shows the sourdough commit; it is experiments~1.` / `git cherry-pick experiments~1`
- Solution: `git cherry-pick experiments~1`. Wrong solution test: `git merge --no-edit experiments` fails with reason (3).

### Level 10 · Moving Camp
- Story: `You built the campfire on top of the scouting trail, but the trail was a dead end. The campfire belongs right on main.`
- Goal: `Move the two campfire commits so they sit directly on top of main, without the trail commit. Leave main itself alone.`
- Setup: `Write camp.txt "base camp\n"`, `Commit "Set up base camp"`, `Branch trail`, `Switch trail`, `Write trail.txt "dead end\n"`, `Commit "Scout the trail"`, `Save trail HEAD`, `Branch campfire`, `Switch campfire`, `Write fire.txt "kindling\n"`, `Commit "Build the campfire"`, `Write fire.txt "kindling\nmarshmallows\n"`, `Commit "Roast marshmallows"`, `Switch main`, `Write tent.txt "pitched\n"`, `Commit "Pitch the tent"`, `Save main HEAD`, `Switch campfire`.
- Check (no main requirement): (1) `Rev("main") == Ref("main")` else `main moved; leave main where it is (gitquest reset starts over)`; (2) `Rev("campfire") != ""` else `the campfire branch is missing`; (3) `!IsAncestor(Ref("trail"), "campfire")` else `campfire still contains the "Scout the trail" commit`; (4) `IsAncestor(Ref("main"), "campfire")` else `campfire does not sit on top of main yet`; (5) `Count("main..campfire") == 2` else `campfire should have exactly 2 commits on top of main`; (6) `Blob("campfire","fire.txt") == "kindling\nmarshmallows\n"` and `!Has("campfire","trail.txt")` else `campfire's files are not right: fire.txt needs both lines and trail.txt must be gone`.
- Hints: `git rebase --onto <new-base> <old-base> <branch> replays only the commits after <old-base>.` / `The new base is main, the old base is trail, and the branch is campfire.` / `git rebase --onto main trail campfire`
- Solution: `git rebase --onto main trail campfire`. Wrong solution test: `git rebase main` (on campfire) replays the trail commit too and fails with reason (5).

### Level 11 · Two Kinds of Forgetting
- Story: `You wrote chapters 2 and 3 as two messy draft commits. Squash them into one commit, but do not lose the edit to chapter 1 that you have not committed yet.`
- Goal: `Replace the two draft commits with one commit named exactly "Write chapters 2 and 3". Keep your uncommitted edit in ch1.txt: do not commit it and do not lose it.`
- Setup: `Write ch1.txt "Chapter 1\n"`, `Commit "Write chapter 1"`, `Save ch1 HEAD`, `Write ch2.txt "Chapter 2\n"`, `Commit "Draft chapter 2"`, `Write ch3.txt "Chapter 3\n"`, `Commit "Draft chapter 3"`, `Save drafts HEAD`, `Write ch1.txt "Chapter 1\nedit: keep me\n"`.
- Check: (1) `File("ch1.txt") == "Chapter 1\nedit: keep me\n"` else `your uncommitted edit to ch1.txt is gone (git reset --hard throws working changes away); run gitquest reset and try --soft`; (2) on main; (3) `Parents("HEAD") == [Ref("ch1")]` else `there should be exactly one commit after "Write chapter 1"`; (4) `Message("HEAD") == "Write chapters 2 and 3"` else `the commit message is "<msg>"`; (5) `Tree("HEAD") == Tree(Ref("drafts"))` else `the new commit should hold ch2.txt and ch3.txt, and the original ch1.txt`.
- Hints: `git reset --soft moves the branch back but keeps the changes staged. git reset --hard throws changes away, even uncommitted ones.` / `git reset --soft HEAD~2 un-commits both drafts and keeps them staged.` / `git reset --soft HEAD~2 && git commit -m "Write chapters 2 and 3" (no -a, so ch1.txt stays uncommitted)`
- Solution: `git reset --soft HEAD~2`; `git commit -m "Write chapters 2 and 3"`. Wrong solution test: `git reset --hard HEAD~2` then the same commit fails with reason (1).

### Level 12 · Secrets in the Attic
- Story: `The attic is full of junk: logs, a cache folder and a file full of passwords. Only your notes belong in the repository.`
- Goal: `Commit a .gitignore that ignores debug.log, the cache/ folder and secrets.env, and commit notes.md. None of the junk may be committed.`
- Setup: `Write attic.txt "dusty\n"`, `Commit "Open the attic"`, `Write debug.log "noise\n"`, `Write cache/a.tmp "tmp\n"`, `Write secrets.env "PASSWORD=swordfish\n"`, `Write notes.md "# Notes\n"`.
- Check: (1) on main; (2) `Has("HEAD",".gitignore")` else `there is no committed .gitignore yet`; (3) for `p` in `debug.log`, `cache/a.tmp`, `secrets.env`: `!Has("HEAD",p)` else `<p> was committed; remove it with git rm --cached <p> and commit again`; (4) `Has("HEAD","notes.md")` else `notes.md is not committed yet`; (5) for the same three `p`: `Ignored(p)` else `<p> is not ignored by your .gitignore`; (6) `!Ignored("notes.md")` else `notes.md is ignored too, but it should be committed`.
- Hints: `A .gitignore file lists patterns for files git should leave alone.` / `The patterns *.log, cache/ and secrets.env match the junk.` / `printf '*.log\ncache/\nsecrets.env\n' > .gitignore && git add .gitignore notes.md && git commit -m "Ignore the junk"`
- Solution: exactly hint 3 (with a real newline-producing `printf`).

## 9. Package layout

```
gitquest/
├── main.go                    // os.Exit(cli.Run(os.Args[1:], cli.RealDeps(version)))
├── internal/
│   ├── gitx/gitx.go           // Git, SetupEnv, CheckEnv, Run, Ok, VersionOK
│   ├── game/root.go           // Root, LevelDir, EnsureRoot, markers, SafeRemove
│   ├── game/progress.go       // StateDir, Progress, Load, Save
│   ├── game/build.go          // Build(root, level)
│   ├── levels/level.go        // Level, Step, Ctx, ApplySteps, All()
│   ├── levels/l01_first_commit.go … l12_gitignore.go
│   └── cli/cli.go
├── .goreleaser.yaml, .github/workflows/{ci.yml,release.yml}
├── README.md
└── LICENSE
```

## 10. Testing requirements

- Tests never touch the real HOME, `~/gitquest`, or real repos: they set `HOME`, `GITQUEST_ROOT`, `GITQUEST_HOME`, `XDG_CONFIG_HOME` to `t.TempDir()` paths and run player commands with `GIT_CONFIG_GLOBAL=<empty temp file>`, `GIT_CONFIG_NOSYSTEM=1`, `GIT_EDITOR=true`.
- Per level: a test that the fresh level fails with its first reason; a test that the intended solution (§8 "Solution") passes; the listed wrong-solution tests.
- Determinism: every level built in two different roots has identical `refs` and `HEAD` commit IDs.
- Safety tests: `SafeRemove` refuses a missing marker, a wrong-level marker, a symlinked level dir, and a root without its marker; `EnsureRoot` refuses a non-empty unmarked dir.
- CI: `gofmt -l .` empty, `go vet ./...`, `go test ./...` on ubuntu-latest and macos-latest (both ship git ≥ 2.32).

## 11. Definition of done (v0.1)

- All commands in §4 behave as specified; tests, vet and gofmt are clean.
- Manual playthrough of levels 1, 4 and 8 in a real shell with `GITQUEST_ROOT=$(mktemp -d)/gq`.
- README: pitch, install, how to play, the 12 level titles, safety notes (game root, markers), `GITQUEST_ROOT`/`GITQUEST_HOME`.
- GoReleaser builds linux/darwin × amd64/arm64.

## 12. Later

- Windows support (replace `/dev/null` with an empty temp config file; test on Windows CI).
- More levels: remotes with a local bare repo, interactive rebase, bisect, worktrees, tags, `git restore`.
- `gitquest diff` to show what the checker expects, achievements, a timer mode.
- Homebrew tap.
