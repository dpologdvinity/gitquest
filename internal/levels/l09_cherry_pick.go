package levels

func init() {
	registry = append(registry, Level{N: 9, Title: "Borrowed Recipe", Story: "The experiments branch has three new recipes. Only the sourdough is good enough for the bakery menu.", Goal: `Copy only the "Perfect the sourdough" commit onto main.`, Hints: [3]string{"git cherry-pick copies a single commit onto the branch you are on.", "git log --oneline experiments shows the sourdough commit; it is experiments~1.", "git cherry-pick experiments~1"}, Setup: []Step{W("menu.txt", "bread\n"), C("Open the bakery"), B("experiments"), S("experiments"), W("glitter.txt", "glitter frosting\n"), C("Try glitter frosting"), W("sourdough.txt", "flour, water, salt, patience\n"), C("Perfect the sourdough"), Sv("sourdough", "HEAD"), W("fishcake.txt", "fish cake\n"), C("Try fish cake"), S("main"), W("sign.txt", "OPEN\n"), C("Paint the sign"), Sv("main", "HEAD")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if !c.IsAncestor(c.Ref("main"), "main") {
			return false, "main lost its own commits"
		}
		if c.Count("--min-parents=2", c.Ref("main")+"..main") != 0 {
			return false, "main got a merge commit; copy just the one commit instead"
		}
		if !c.Has("main", "sourdough.txt") || c.Blob("main", "sourdough.txt") != c.Blob(c.Ref("sourdough"), "sourdough.txt") {
			return false, "main has no sourdough.txt yet"
		}
		if c.Has("main", "glitter.txt") || c.Has("main", "fishcake.txt") {
			return false, "main picked up recipes it should not have"
		}
		return true, ""
	}})
}
