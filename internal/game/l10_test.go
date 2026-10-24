package game

func init() {
	plays = append(plays, play{n: 10, fresh: `campfire still contains the "Scout the trail" commit`, solve: `git rebase -q --onto main trail campfire`, wrong: `git rebase -q main`, wrongReason: "campfire should have exactly 2 commits on top of main"})
}
