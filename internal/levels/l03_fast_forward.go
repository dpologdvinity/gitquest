package levels

func init() {
	registry = append(registry, Level{N: 3, Title: "Catching Up", Story: "Your crew patched the raft on the repairs branch. The main branch still points at the old, leaky raft.", Goal: "Bring main up to date with repairs without creating a merge commit.", Hints: [3]string{"When main has no new commits of its own, a merge can just slide main forward.", "Stay on main and merge the repairs branch.", "git merge --ff-only repairs"}, Setup: []Step{W("raft.txt", "logs\n"), C("Build the raft"), B("repairs"), S("repairs"), W("sail.txt", "patched\n"), C("Patch the sail"), W("knots.txt", "tight\n"), C("Tie the knots"), Sv("repairs", "repairs"), S("main")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if len(c.Parents("main")) >= 2 {
			return false, "main got a merge commit; this merge can be a fast-forward"
		}
		if c.Rev("main") != c.Ref("repairs") {
			return false, "main does not point at the same commit as repairs yet"
		}
		return true, ""
	}})
}
