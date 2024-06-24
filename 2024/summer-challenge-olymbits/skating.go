package main

import "strings"

type Skating struct {
	Race
}

func NewSkating(skaters ...*Skater) *Skating {
	s := Skating{}

	j := 0
	for i := range skaters {
		skaters[i].regs = [2]*int{&s.regs[j], &s.regs[j+3]}
		j++
		s.players[i] = skaters[i]
	}

	return &s
}

func (s Skating) turnsLeft() int {
	return s.regs[6]
}

var Ranks = [4][2]int{
	{1, -1},
	{2, 0},
	{2, 1},
	{3, 2},
}

func (s Skating) rank(cmd Command) [2]int {
	i := strings.IndexRune(s.gpu, rune(cmd[0]))

	return Ranks[i]
}
func (s Skating) evalCollision(pos, playerIdx int) float64 {
	pos = pos % 10
	score := 0
	for i, p := range s.players {
		p := p.(*Skater)
		if p.risk() < 0 {
			if pos != p.spaces()%10 {
				continue
			}

			score += 2
		}

		if i != playerIdx {
			// Estimate opponent's possible new positions
			oppPositions := []int{
				(p.spaces() + 1) % 10,
				(p.spaces() + 2) % 10,
				(p.spaces() + 2) % 10,
				(p.spaces() + 3) % 10,
			}

			for _, oppPos := range oppPositions {
				if pos != oppPos {
					continue
				}

				score++
			}
		}
	}
	return 4 * normalize(float64(score), 0, 8)
}

func (s Skating) skater(idx int) *Skater {
	return s.Player(idx).(*Skater)
}

func (s Skating) Place(p Player) int {
	place := 1

	for _, player := range s.players {
		if player == p {
			continue
		}

		player := player.(*Skater)
		p := p.(*Skater)
		if player.spaces() > p.spaces() {
			place++
		}
	}

	return place
}

func (s Skating) Eval(cmd Command, playerIdx int) int {
	if s.isEOG() {
		return 0
	}

	player := s.skater(playerIdx)

	if player.risk() < 0 {
		return 0
	}

	rank := s.rank(cmd)
	deltaSpaces := rank[0]
	deltaRisk := rank[1]

	collision := s.evalCollision((player.spaces() + deltaSpaces), playerIdx)

	score := 0.0
	if risk := deltaRisk + player.risk(); risk > 4 {
		score -= float64(risk)
		score += float64(deltaSpaces)
	} else {
		score += float64(deltaSpaces)
		score -= collision
		score -= float64(player.risk()-deltaRisk) * 0.25
	}

	return s.normalize(float64(score), -3, 4)
}

type Skater struct {
	Contestant
}

func NewSkater() *Skater {
	return &Skater{}
}

func (s Skater) spaces() int {
	return *s.regs[0]
}
func (s Skater) risk() int {
	return *s.regs[1]
}
