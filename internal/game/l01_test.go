package game

func init() {
	plays = append(plays, play{n: 1, fresh: "there are no commits yet", solve: `git add logbook.txt && git commit -q -m "First entry"`})
}
