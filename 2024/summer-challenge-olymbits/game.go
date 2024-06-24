package main

import "math"

type Game interface {
	Place(Player) int

	Player(idx int) Player

	// Update state of the game
	Update(gpu string, regs [7]int)

	// Eval rates simulated move for the player ranged from 0 to 100
	Eval(cmd Command, playerIdx int) int

	// isEOG checks if the current session has ended
	isEOG() bool
}

type Race struct {
	gpu     string
	regs    [7]int
	players [3]Player
}

func (r Race) normalize(n, min, max float64) int {
	return int(math.Ceil(100 * normalize(n, min, max)))
}

func (r *Race) Update(gpu string, regs [7]int) {
	r.gpu = gpu
	r.regs = regs
}

func (r Race) Player(idx int) Player {
	return r.players[idx]
}

func (r Race) isEOG() bool {
	return r.gpu == EOG
}
