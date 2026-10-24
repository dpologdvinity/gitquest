package levels

import "fmt"

func init() {
	registry = append(registry, Level{N: 11, Title: "Two Kinds of Forgetting", Story: "You wrote chapters 2 and 3 as two messy draft commits. Squash them into one commit, but do not lose the edit to chapter 1 that you have not committed yet.", Goal: `Replace the two draft commits with one commit named exactly "Write chapters 2 and 3". Keep your uncommitted edit in ch1.txt: do not commit it and do not lose it.`, Hints: [3]string{"git reset --soft moves the branch back but keeps the changes staged. git reset --hard throws changes away, even uncommitted ones.", "git reset --soft HEAD~2 un-commits both drafts and keeps them staged.", `git reset --soft HEAD~2 && git commit -m "Write chapters 2 and 3" (no -a, so ch1.txt stays uncommitted)`}, Setup: []Step{W("ch1.txt", "Chapter 1\n"), C("Write chapter 1"), Sv("ch1", "HEAD"), W("ch2.txt", "Chapter 2\n"), C("Draft chapter 2"), W("ch3.txt", "Chapter 3\n"), C("Draft chapter 3"), Sv("drafts", "HEAD"), W("ch1.txt", "Chapter 1\nedit: keep me\n")}, Check: func(c *Ctx) (bool, string) {
		if s, ok := c.File("ch1.txt"); !ok || s != "Chapter 1\nedit: keep me\n" {
			return false, "your uncommitted edit to ch1.txt is gone (git reset --hard throws working changes away); run gitquest reset and try --soft"
		}
		if !c.OnMaster() {
			return false, "switch back to " + c.PrimaryBranch() + " first"
		}
		p := c.Parents("HEAD")
		if len(p) != 1 || p[0] != c.Ref("ch1") {
			return false, `there should be exactly one commit after "Write chapter 1"`
		}
		if msg := c.Message("HEAD"); msg != "Write chapters 2 and 3" {
			return false, fmt.Sprintf("the commit message is %q", msg)
		}
		if c.Tree("HEAD") != c.Tree(c.Ref("drafts")) {
			return false, "the new commit should hold ch2.txt and ch3.txt, and the original ch1.txt"
		}
		return true, ""
	}})
}
