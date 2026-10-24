package game

func init() {
	plays = append(plays, play{n: 6, fresh: "nothing has been undone yet", solve: `git revert --no-edit HEAD~1`, wrong: `git reset -q --hard HEAD~2`, wrongReason: "history was rewritten, but the village already has those commits; run gitquest reset and try git revert"})
}
