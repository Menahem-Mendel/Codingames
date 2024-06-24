package main

import (
	"bufio"
	"fmt"
	"math"
)

const (
	ARCHERY  = "ARCHERY"
	HURDLING = "HURDLING"
	SKATING  = "SKATING"
	DIVING   = "DIVING"
)
const EOG = "GAME_OVER"

type Command string

const (
	LEFT  Command = "LEFT"
	DOWN  Command = "DOWN"
	RIGHT Command = "RIGHT"
	UP    Command = "UP"
)

type Medal int

const (
	GOLD   Medal = 0
	SILVER Medal = 1
	BRONZE Medal = 2
)

func normalize(n, min, max float64) float64 {
	return 2*((n-min)/(max-min)) - 1
}

func UpdateGame(g Game, gpu string, regs [7]int) {
	g.Update(gpu, regs)
}

func UpdatePlayer(p Player, score Score) {
	p.Update(score)
}

func ParseState(scanner *bufio.Scanner) (gpu string, regs [7]int) {
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &gpu,
		&regs[0], &regs[1], &regs[2], &regs[3], &regs[4], &regs[5], &regs[6])

	return
}

func toInt(str string) int {
	var result int
	fmt.Sscan(str, &result)
	return result
}

func checkForHurdle(track string, distance int) bool {
	return len(track) > distance && track[distance] == HURDLE
}

func calcScore(remained, move int) int {
	if remained < move && remained > 0 {
		return remained
	}
	return move
}

func clamp(a, min, max float64) float64 {
	if a > max {
		a = max
	} else if a < min {
		a = min
	}
	return a
}

func dist(a, b Coord) float64 {
	return math.Sqrt(float64((a.x-b.x)*(a.x-b.x)) + float64((a.y-b.y)*(a.y-b.y)))
}
