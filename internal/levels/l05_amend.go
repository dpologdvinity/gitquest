package levels

import "fmt"

func init() {
	registry = append(registry, Level{N: 5, Title: "Typo in the Telegram", Story: "You just sent a telegram to HQ, but the commit message has a typo. Nobody has seen it yet, so you can fix it in place.", Goal: `Change the last commit's message to exactly "Send telegram" without adding a new commit.`, Hints: [3]string{"You can rewrite the most recent commit as long as nobody else has it yet.", "git commit --amend replaces the last commit with a new one.", `git commit --amend -m "Send telegram"`}, Setup: []Step{W("header.txt", "TO: HQ\n"), C("Add the header"), Sv("base", "HEAD"), W("telegram.txt", "ARRIVING TUESDAY STOP\n"), C("Send telgram"), Sv("typo", "HEAD")}, Check: func(c *Ctx) (bool, string) {
		if !c.OnMain() {
			return false, "switch back to main first"
		}
		if c.Rev("HEAD") == c.Ref("typo") {
			return false, `the last commit still says "Send telgram"`
		}
		p := c.Parents("HEAD")
		if len(p) != 1 || p[0] != c.Ref("base") {
			return false, `the fixed commit should sit right on top of "Add the header"; amend instead of adding a commit`
		}
		if msg := c.Message("HEAD"); msg != "Send telegram" {
			return false, fmt.Sprintf(`the message is %q, not "Send telegram"`, msg)
		}
		if c.Tree("HEAD") != c.Tree(c.Ref("typo")) {
			return false, "the files changed too; only the message should change"
		}
		return true, ""
	}})
}
