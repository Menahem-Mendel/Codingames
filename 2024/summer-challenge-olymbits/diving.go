package main

const (
	L = 'l'
	D = 'D'
	R = 'R'
	U = 'U'
)

type Diving struct {
	Race
}

func NewDiving(divers ...*Diver) *Diving {
	d := Diving{}

	j := 0
	for i := range divers {
		divers[i].regs = [2]*int{&d.regs[j], &d.regs[j+3]}
		j++
		d.players[i] = divers[i]
	}

	return &d
}

func (d Diving) diver(idx int) *Diver {
	return d.Player(idx).(*Diver)
}

func (d Diving) Place(p Player) int {
	place := 1

	for _, player := range d.players {
		if player == p {
			continue
		}

		player := player.(*Diver)
		p := p.(*Diver)
		if player.points()+player.combo() > p.points() {
			place++
		}
	}

	return place
}

func (d Diving) Eval(cmd Command, playerIdx int) int {
	if d.isEOG() {
		return 0
	}

	player := d.diver(playerIdx)

	score := 0
	if len(d.gpu) > 0 && d.gpu[0] != cmd[0] {
		score = -player.combo()
	} else {
		score = player.combo() + 1
	}

	// max := player.combo()*len(d.gpu) + (len(d.gpu)*(len(d.gpu)-1))/2 + 1
	max := player.combo() + 1
	min := -player.combo()
	return d.normalize(float64(score), float64(min), float64(max))
}

type Diver struct {
	Contestant
}

func NewDiver() *Diver {
	return &Diver{}
}

func (d Diver) points() int {
	return *d.regs[0]
}

func (d Diver) combo() int {
	return *d.regs[1]
}
