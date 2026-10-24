package levels

import "fmt"

func init() {
	registry = append(registry, Level{N: 12, Title: "Secrets in the Attic", Story: "The attic is full of junk: logs, a cache folder and a file full of passwords. Only your notes belong in the repository.", Goal: "Commit a .gitignore that ignores debug.log, the cache/ folder and secrets.env, and commit notes.md. None of the junk may be committed.", Hints: [3]string{"A .gitignore file lists patterns for files git should leave alone.", "The patterns *.log, cache/ and secrets.env match the junk.", `printf '*.log\ncache/\nsecrets.env\n' > .gitignore && git add .gitignore notes.md && git commit -m "Ignore the junk"`}, Setup: []Step{W("attic.txt", "dusty\n"), C("Open the attic"), W("debug.log", "noise\n"), W("cache/a.tmp", "tmp\n"), W("secrets.env", "PASSWORD=swordfish\n"), W("notes.md", "# Notes\n")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if !c.Has("HEAD", ".gitignore") {
			return false, "there is no committed .gitignore yet"
		}
		paths := []string{"debug.log", "cache/a.tmp", "secrets.env"}
		for _, p := range paths {
			if c.Has("HEAD", p) {
				return false, fmt.Sprintf("%s was committed; remove it with git rm --cached %s and commit again", p, p)
			}
		}
		if !c.Has("HEAD", "notes.md") {
			return false, "notes.md is not committed yet"
		}
		for _, p := range paths {
			if !c.Ignored(p) {
				return false, p + " is not ignored by your .gitignore"
			}
		}
		if c.Ignored("notes.md") {
			return false, "notes.md is ignored too, but it should be committed"
		}
		return true, ""
	}})
}
