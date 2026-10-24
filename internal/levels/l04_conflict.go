package levels

import "strings"

func init() {
	registry = append(registry, Level{N: 4, Title: "Two Captains", Story: "Two captains wrote different courses into orders.txt. The ship can only follow one set of orders, so you must merge them.", Goal: "Merge captain-b into master. Resolve the conflict so orders.txt mentions both west and east, then finish the merge.", Hints: [3]string{"Start with git merge captain-b. Git stops and marks the conflict inside orders.txt.", "Edit orders.txt: delete the <<<<<<<, ======= and >>>>>>> lines and write a course that has both west and east.", "git merge captain-b, edit orders.txt, then git add orders.txt && git commit --no-edit"}, Setup: []Step{W("orders.txt", "Course: north\n"), C("Agree on a course"), B("captain-b"), W("orders.txt", "Course: west\n"), C("Captain A: sail west"), Sv("master", "master"), S("captain-b"), W("orders.txt", "Course: east\n"), C("Captain B: sail east"), Sv("b", "captain-b"), S("master")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMaster() {
			return false, "switch back to " + c.PrimaryBranch() + " first"
		}
		if c.Merging() {
			return false, "the merge is still in progress: fix orders.txt, git add it, then git commit"
		}
		p := c.Parents("HEAD")
		if len(p) != 2 || p[0] != c.Ref("master") || p[1] != c.Ref("b") {
			return false, "master is not a merge of both captains' orders yet"
		}
		blob := c.Blob("HEAD", "orders.txt")
		for _, line := range strings.Split(blob, "\n") {
			if strings.HasPrefix(line, "<<<<<<<") || strings.HasPrefix(line, "=======") || strings.HasPrefix(line, ">>>>>>>") {
				return false, "orders.txt still has conflict markers in it"
			}
		}
		low := strings.ToLower(blob)
		if !strings.Contains(low, "west") || !strings.Contains(low, "east") {
			return false, "orders.txt should mention both west and east"
		}
		return true, ""
	}})
}
