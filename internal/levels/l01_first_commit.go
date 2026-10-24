package levels

func init() {
	registry = append(registry, Level{N: 1, Title: "The Empty Logbook", Story: "You are the new keeper of the gitquest lighthouse. The old keeper left a logbook on the desk, but nothing in it was ever saved.", Goal: "Save logbook.txt in a commit on the main branch.", Hints: [3]string{"A commit is a saved snapshot. First tell git which files belong in it.", "git add logbook.txt stages the file for the next commit.", `git add logbook.txt && git commit -m "First entry"`}, Setup: []Step{W("logbook.txt", "Day 1: the lamp is lit.\n")}, Check: func(c *Ctx) (bool, string) {
		if c.Rev("HEAD") == "" {
			return false, "there are no commits yet"
		}
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if !c.Has("HEAD", "logbook.txt") {
			return false, "the latest commit does not contain logbook.txt"
		}
		return true, ""
	}})
}
