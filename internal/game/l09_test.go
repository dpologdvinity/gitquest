package game

func init() {
	plays = append(plays, play{n: 9, fresh: "main has no sourdough.txt yet", solve: `git cherry-pick experiments~1`, wrong: `git merge -q --no-edit experiments`, wrongReason: "main got a merge commit; copy just the one commit instead"})
}
