gitquest: learn git by escaping puzzle rooms
============================================

`gitquest` is a 12-level escape-room game played with real Git repositories. Each room starts with a small story and a Git puzzle. Use Git in your shell, then ask `gitquest` to check your solution.

## Install

You need Git 2.32 or newer. Install with Go:

```sh
go install github.com/dpologdvinity/gitquest@latest
```

Tagged releases provide pre-compiled archives for Linux and macOS (amd64 and arm64) on the project's GitHub Releases page. Extract the `gitquest` executable and place it on your `PATH`.

## How to play

Run `gitquest start`, change into the level folder it prints, solve the puzzle with normal `git` commands, then run `gitquest check`. Use `gitquest hint` for up to three progressive hints. `gitquest status` lists all rooms, `gitquest level N` opens any room, and `gitquest reset [N]` rebuilds a room from scratch.

Here is a level 1 session:

```text
$ gitquest start
Level 1/12 · The Empty Logbook

You are the new keeper of the gitquest lighthouse. The old keeper left a logbook on the desk, but nothing in it was ever saved.

Goal: Save logbook.txt in a commit on the main branch.

Your level is ready:
  cd ~/gitquest/level-01
Run gitquest check when you think you're done, or gitquest hint if you're stuck.
$ cd ~/gitquest/level-01
$ gitquest check
✗ Not solved yet: there are no commits yet
Stuck? Try gitquest hint
$ git add logbook.txt
$ git commit -q -m 'First entry'
$ gitquest check
✓ Level 1 solved: The Empty Logbook
Next: gitquest level 2
```

## Levels

| # | Room | Git concept |
|---|---|---|
| 1 | The Empty Logbook | First commit |
| 2 | A Fork in the Path | Branch and switch |
| 3 | Catching Up | Fast-forward merge |
| 4 | Two Captains | Merge conflict |
| 5 | Typo in the Telegram | Amend |
| 6 | Undo the Spell | Revert |
| 7 | Pocket Dimension | Stash |
| 8 | The Lost Treasure Map | Reflog rescue |
| 9 | Borrowed Recipe | Cherry-pick |
| 10 | Moving Camp | `rebase --onto` |
| 11 | Two Kinds of Forgetting | Soft and hard reset |
| 12 | Secrets in the Attic | `.gitignore` |

## Architecture Overview

`gitquest` is built purely with the Go standard library and the host system's `git` binary (version 2.32+).
- **`internal/gitx`**: Manages child process execution with sanitized environments, isolating author metadata, dates, and configuration paths.
- **`internal/game`**: Enforces strict directory isolation using `.gitquest-root` and `.gitquest-level` sentinels, guarantees atomic progress tracking in `progress.json`, and coordinates deterministic level builds.
- **`internal/levels`**: Contains level specifications and deterministic step definitions. Verifications inspect Git plumbing (`rev-parse`, `symbolic-ref`, `cat-file`, `merge-base`) without parsing porcelain output.
- **`internal/cli`**: Wires command dispatch, player progress reporting, and user hints.

For comprehensive architectural design and technical details, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Safety and locations

Level repositories live under `~/gitquest` by default. Set `GITQUEST_ROOT` to an absolute path to put them elsewhere. Gitquest marks its root with `.gitquest-root` and every level with `.gitquest-level`. It refuses to use an existing nonempty folder without the root marker or reset a level without its matching marker. `gitquest reset` deletes and rebuilds a marked level, including any work you made there.

Progress is stored in `$GITQUEST_HOME/progress.json` when `GITQUEST_HOME` is set, otherwise under `$XDG_STATE_HOME/gitquest` or `~/.local/state/gitquest`.

Each level repository has its own Git identity (`Adventurer <adventurer@gitquest.invalid>`) and disables hooks. Gitquest does not touch repositories outside its game root.

## Development & Contributing

To develop, test, or build `gitquest` locally:

```bash
git clone https://github.com/dpologdvinity/gitquest.git
cd gitquest

# Run full test suite
go test -v ./...

# Run static analysis and linters
test -z "$(gofmt -l .)"
go vet ./...
golangci-lint run ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines on coding standards, tests, and pull request conventions.

## License

MIT. See [LICENSE](LICENSE).
