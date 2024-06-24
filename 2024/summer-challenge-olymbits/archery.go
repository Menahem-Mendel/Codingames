package main

import "math"

type Archery struct {
	Race
}

func NewArchery(archers ...*Archer) *Archery {
	a := Archery{}

	j := 0
	for i := range archers {
		archers[i].regs = [2]*int{&a.regs[j], &a.regs[j+1]}
		j += 2
		a.players[i] = archers[i]
	}

	return &a
}

func (a Archery) wind() int {
	return toInt(string(a.gpu[0]))
}

func (a Archery) Place(p Player) int {
	place := 1

	for _, player := range a.players {
		if player == p {
			continue
		}

		player := player.(*Archer)
		p := p.(*Archer)
		if dist(player.coord(), Origin) > dist(p.coord(), Origin) {
			place++
		}
	}

	return place
}

func (a Archery) Eval(cmd Command, playerIdx int) int {
	if a.isEOG() {
		return 0
	}

	player := a.Player(playerIdx).(*Archer)

	score := 0.0

	coord := player.coord()

	switch cmd {
	case LEFT:
		coord.x -= a.wind()
	case UP:
		coord.y -= a.wind()
	case RIGHT:
		coord.x += a.wind()
	case DOWN:
		coord.y += a.wind()
	}

	coord.x = int(clamp(float64(coord.x), -20, 20))
	coord.y = int(clamp(float64(coord.y), -20, 20))

	delta := math.Sqrt(float64(player.coord().x*player.coord().x + player.coord().y*player.coord().y))
	deltaEval := math.Sqrt(float64(coord.x*coord.x + coord.y*coord.y))

	score = delta - deltaEval
	return int(100 * normalize(float64(score), -9, 9+float64(len(a.gpu)-1)))
}

var Origin Coord = Coord{0, 0}

type Coord struct {
	x int // x coordinate
	y int // y coordinate
}

type Archer struct {
	Contestant
}

func NewArcher() *Archer {
	return &Archer{}
}

func (a Archer) coord() Coord {
	return Coord{
		x: *a.regs[0],
		y: *a.regs[1],
	}
}
