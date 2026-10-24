package levels

func init() {
	registry = append(registry, Level{N: 6, Title: "Undo the Spell", Story: "Someone cast a growth spell on the village garden, and everyone already pulled it. You must undo it without rewriting shared history.", Goal: `Undo the "Cast a growth spell" commit with a new commit, keeping every existing commit.`, Hints: [3]string{"Shared history should not be rewritten. Add a commit that does the opposite instead.", "git log --oneline shows the commit to undo; it is HEAD~1.", "git revert --no-edit HEAD~1"}, Setup: []Step{W("garden.txt", "roses\n"), C("Plant the garden"), W("spell.txt", "grow grow grow\n"), C("Cast a growth spell"), W("garden.txt", "roses\nwatered\n"), C("Water the plants"), Sv("tip", "HEAD")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if !c.IsAncestor(c.Ref("tip"), "HEAD") {
			return false, "history was rewritten, but the village already has those commits; run gitquest reset and try git revert"
		}
		if c.Rev("HEAD") == c.Ref("tip") {
			return false, "nothing has been undone yet"
		}
		if c.Has("HEAD", "spell.txt") {
			return false, "spell.txt is still in the latest commit"
		}
		if c.Blob("HEAD", "garden.txt") != "roses\nwatered\n" {
			return false, "garden.txt lost its watering; undo only the spell"
		}
		return true, ""
	}})
}
