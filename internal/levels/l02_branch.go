package levels

func init() {
	registry = append(registry, Level{N: 2, Title: "A Fork in the Path", Story: "Behind the lighthouse the path splits in two. You want to explore the new trail without scribbling on the master map.", Goal: "Create a branch named explore and switch to it.", Hints: [3]string{"Branches are movable name tags for commits.", "git branch explore creates it; git switch explore moves you onto it.", "git switch -c explore"}, Setup: []Step{W("map.txt", "lighthouse\n"), C("Draw the map")}, Check: func(c *Ctx) (bool, string) {
		if c.Rev("refs/heads/explore") == "" {
			return false, "there is no branch named explore"
		}
		if c.Head() != "refs/heads/explore" {
			return false, "the explore branch exists, but you are not on it"
		}
		return true, ""
	}})
}
