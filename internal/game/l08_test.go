package game

func init() {
	plays = append(plays, play{n: 8, fresh: "there is no branch named treasure", solve: `git branch treasure HEAD@{1}`, wrong: `git branch treasure`, wrongReason: "treasure exists, but it does not contain the lost map commit"})
}
