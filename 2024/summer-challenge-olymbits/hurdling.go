package main

const (
	DOT    = '.'
	HURDLE = '#'
)

type Hurdling struct {
	Race
}

func NewHurdling(hurdlers ...*Hurdler) *Hurdling {
	h := Hurdling{}

	j := 0
	for i := range hurdlers {
		hurdlers[i].regs = [2]*int{&h.regs[j], &h.regs[j+3]}
		j++

		h.players[i] = hurdlers[i]
	}

	return &h
}

func (h Hurdling) willStun(cmd Command, pos int) bool {
	if len(h.gpu) < pos {
		return false
	}

	res := false
	for i := 1; i <= Steps[cmd]; i++ {
		res = res || checkForHurdle(h.gpu[pos:], i)
	}

	return res
}

func (h Hurdling) calcMove(track string, move int) int {
	score := move
	for i := 0; i < move && i < len(track); i++ {
		if track[i] == HURDLE {
			score = i
			break
		}
	}

	return score
}

func (h Hurdling) Place(p Player) int {
	place := 1

	for _, player := range h.players {
		if player == p {
			continue
		}

		player := player.(*Hurdler)
		p := p.(*Hurdler)
		if player.pos() > p.pos() {
			place++
		}
	}

	return place
}

var Steps = map[Command]int{
	LEFT:  1,
	UP:    2,
	DOWN:  2,
	RIGHT: 3,
}

func (h Hurdling) Eval(cmd Command, playerIdx int) int {
	if h.isEOG() {
		return 0
	}

	player := h.players[playerIdx].(*Hurdler)

	if player.stuns() > 0 {
		return 0
	}

	score := 0

	if h.willStun(cmd, player.pos()) {
		score -= 3
	}

	score += h.calcMove(h.gpu[player.pos():], Steps[cmd])

	return h.normalize(float64(score), -2, 3)
}

type Hurdler struct {
	Contestant
}

func NewHurdler() *Hurdler {
	return &Hurdler{}
}

func (h Hurdler) pos() int {
	return *h.regs[0]
}

func (h Hurdler) stuns() int {
	return *h.regs[1]
}
