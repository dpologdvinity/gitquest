package levels

func init() {
	registry = append(registry, Level{N: 7, Title: "Pocket Dimension", Story: "You were drafting plan B when a messenger burst in: the bridge is broken! Your draft is not ready to commit, and git will not let you switch to the hotfix branch with it lying around.", Goal: "Put your draft aside, commit a file named bridge.txt on the hotfix branch, then return to main and bring your draft back (uncommitted).", Hints: [3]string{"git stash tucks uncommitted changes away and leaves a clean working tree.", "Stash, switch to hotfix, add and commit bridge.txt, switch back to main, then bring the stash back.", `git stash && git switch hotfix && echo fixed > bridge.txt && git add bridge.txt && git commit -m "Fix the bridge" && git switch main && git stash pop`}, Setup: []Step{W("notes.txt", "plan A\n"), C("Write plan A"), B("hotfix"), S("hotfix"), W("notes.txt", "plan A\nbridge: broken\n"), C("Report the broken bridge"), Sv("hotfix", "hotfix"), S("main"), Sv("main", "main"), W("notes.txt", "plan A\nplan B (draft)\n")}, Check: func(c *Ctx) (bool, string) {
		if c.Rev("hotfix") == c.Ref("hotfix") {
			return false, "hotfix has no new commit yet"
		}
		if !c.IsAncestor(c.Ref("hotfix"), "hotfix") {
			return false, "the hotfix branch lost its earlier commit"
		}
		if !c.Has("hotfix", "bridge.txt") {
			return false, "the newest hotfix commit has no bridge.txt"
		}
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if c.Rev("main") != c.Ref("main") {
			return false, "main changed; your draft should stay uncommitted"
		}
		if s, ok := c.File("notes.txt"); !ok || s != "plan A\nplan B (draft)\n" {
			return false, "your plan B draft is not back in notes.txt yet"
		}
		return true, ""
	}})
}
