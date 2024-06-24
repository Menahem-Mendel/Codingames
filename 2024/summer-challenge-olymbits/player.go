package main

type Player interface {
	// Update state of the player
	Update(score Score)

	// Score earned by player, calculated based on the formula
	// 3*gold + silver medals
	Score() int
}

type Contestant struct {
	regs  [2]*int
	score Score
}

// Score earned in mini-game
// calculated by formula: 3*gold + silver
func (c Contestant) Score() int {
	return c.score.Calc()
}

func (c *Contestant) Update(score Score) {
	c.score = score
}
