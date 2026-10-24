package levels

func init() {
	registry = append(registry, Level{N: 3, Title: "Catching Up", Story: "Your crew patched the raft on the repairs branch. The master branch still points at the old, leaky raft.", Goal: "Bring master up to date with repairs without creating a merge commit.", Hints: [3]string{"When master has no new commits of its own, a merge can just slide master forward.", "Stay on master and merge the repairs branch.", "git merge --ff-only repairs"}, Setup: []Step{W("raft.txt", "logs\n"), C("Build the raft"), B("repairs"), S("repairs"), W("sail.txt", "patched\n"), C("Patch the sail"), W("knots.txt", "tight\n"), C("Tie the knots"), Sv("repairs", "repairs"), S("master")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMaster() {
			return false, "switch back to " + c.PrimaryBranch() + " first"
		}
		if len(c.Parents(c.PrimaryBranch())) >= 2 {
			return false, "master got a merge commit; this merge can be a fast-forward"
		}
		if c.Rev(c.PrimaryBranch()) != c.Ref("repairs") {
			return false, "master does not point at the same commit as repairs yet"
		}
		return true, ""
	}})
}
