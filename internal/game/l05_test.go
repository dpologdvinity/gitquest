package game

func init() {
	plays = append(plays, play{n: 5, fresh: `the last commit still says "Send telgram"`, solve: `git commit -q --amend -m "Send telegram"`, wrong: `git commit -q --allow-empty -m "Send telegram"`, wrongReason: `the fixed commit should sit right on top of "Add the header"; amend instead of adding a commit`})
}
