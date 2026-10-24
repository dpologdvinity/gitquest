package game

func init() {
	plays = append(plays, play{n: 3, fresh: "master does not point at the same commit as repairs yet", solve: `git merge -q --ff-only repairs`, wrong: `git merge -q --no-ff --no-edit repairs`, wrongReason: "master got a merge commit; this merge can be a fast-forward"})
}
