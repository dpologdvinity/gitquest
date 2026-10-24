package levels

func init() {
	registry = append(registry, Level{N: 9, Title: "Borrowed Recipe", Story: "The experiments branch has three new recipes. Only the sourdough is good enough for the bakery menu.", Goal: `Copy only the "Perfect the sourdough" commit onto master.`, Hints: [3]string{"git cherry-pick copies a single commit onto the branch you are on.", "git log --oneline experiments shows the sourdough commit; it is experiments~1.", "git cherry-pick experiments~1"}, Setup: []Step{W("menu.txt", "bread\n"), C("Open the bakery"), B("experiments"), S("experiments"), W("glitter.txt", "glitter frosting\n"), C("Try glitter frosting"), W("sourdough.txt", "flour, water, salt, patience\n"), C("Perfect the sourdough"), Sv("sourdough", "HEAD"), W("fishcake.txt", "fish cake\n"), C("Try fish cake"), S("master"), W("sign.txt", "OPEN\n"), C("Paint the sign"), Sv("master", "HEAD")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMaster() {
			return false, "switch back to " + c.PrimaryBranch() + " first"
		}
		if !c.IsAncestor(c.Ref("master"), c.PrimaryBranch()) {
			return false, "master lost its own commits"
		}
		if c.Count("--min-parents=2", c.Ref("master")+".."+c.PrimaryBranch()) != 0 {
			return false, "master got a merge commit; copy just the one commit instead"
		}
		if !c.Has(c.PrimaryBranch(), "sourdough.txt") || c.Blob(c.PrimaryBranch(), "sourdough.txt") != c.Blob(c.Ref("sourdough"), "sourdough.txt") {
			return false, "master has no sourdough.txt yet"
		}
		if c.Has(c.PrimaryBranch(), "glitter.txt") || c.Has(c.PrimaryBranch(), "fishcake.txt") {
			return false, "master picked up recipes it should not have"
		}
		return true, ""
	}})
}
