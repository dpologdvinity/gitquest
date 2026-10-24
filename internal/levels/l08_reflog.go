package levels

func init() {
	registry = append(registry, Level{N: 8, Title: "The Lost Treasure Map", Story: "Last night you drew a treasure map, then panicked and ran git reset --hard. The commit looks gone, but git quietly remembers everywhere HEAD has been.", Goal: "Find the lost commit and create a branch named treasure that points at it.", Hints: [3]string{"git reflog lists every commit HEAD has pointed at, even after a reset.", `Find "Draw the treasure map" in git reflog. HEAD@{1} is where HEAD was before the reset.`, "git branch treasure HEAD@{1}"}, Setup: []Step{W("journal.txt", "set sail\n"), C("Start the journal"), W("treasure-map.txt", "X marks the spot\n"), C("Draw the treasure map"), Sv("lost", "HEAD"), RH("HEAD~1")}, Check: func(c *Ctx) (bool, string) {
		if c.Rev("refs/heads/treasure") == "" {
			return false, "there is no branch named treasure"
		}
		if !c.IsAncestor(c.Ref("lost"), "refs/heads/treasure") {
			return false, "treasure exists, but it does not contain the lost map commit"
		}
		return true, ""
	}})
}
