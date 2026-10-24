package levels

func init() {
	registry = append(registry, Level{N: 10, Title: "Moving Camp", Story: "You built the campfire on top of the scouting trail, but the trail was a dead end. The campfire belongs right on main.", Goal: "Move the two campfire commits so they sit directly on top of main, without the trail commit. Leave main itself alone.", Hints: [3]string{"git rebase --onto <new-base> <old-base> <branch> replays only the commits after <old-base>.", "The new base is main, the old base is trail, and the branch is campfire.", "git rebase --onto main trail campfire"}, Setup: []Step{W("camp.txt", "base camp\n"), C("Set up base camp"), B("trail"), S("trail"), W("trail.txt", "dead end\n"), C("Scout the trail"), Sv("trail", "HEAD"), B("campfire"), S("campfire"), W("fire.txt", "kindling\n"), C("Build the campfire"), W("fire.txt", "kindling\nmarshmallows\n"), C("Roast marshmallows"), S("main"), W("tent.txt", "pitched\n"), C("Pitch the tent"), Sv("main", "HEAD"), S("campfire")}, Check: func(c *Ctx) (bool, string) {
		if c.Rev("main") != c.Ref("main") {
			return false, "main moved; leave main where it is (gitquest reset starts over)"
		}
		if c.Rev("campfire") == "" {
			return false, "the campfire branch is missing"
		}
		if c.IsAncestor(c.Ref("trail"), "campfire") {
			return false, `campfire still contains the "Scout the trail" commit`
		}
		if !c.IsAncestor(c.Ref("main"), "campfire") {
			return false, "campfire does not sit on top of main yet"
		}
		if c.Count("main..campfire") != 2 {
			return false, "campfire should have exactly 2 commits on top of main"
		}
		if c.Blob("campfire", "fire.txt") != "kindling\nmarshmallows\n" || c.Has("campfire", "trail.txt") {
			return false, "campfire's files are not right: fire.txt needs both lines and trail.txt must be gone"
		}
		return true, ""
	}})
}
