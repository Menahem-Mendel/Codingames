package main

import (
	"os"
	"strings"
	"math"
	"bufio"
	"fmt"
)

const (
	ARCHERY			= "ARCHERY"
	DOWN		Command	= "DOWN"
	SKATING			= "SKATING"
	LEFT		Command	= "LEFT"
	SILVER		Medal	= 1
	BRONZE		Medal	= 2
	DOT			= '.'
	HURDLE			= '#'
	EOG			= "GAME_OVER"
	GOLD		Medal	= 0
	L			= 'l'
	U			= 'U'
	HURDLING		= "HURDLING"
	DIVING			= "DIVING"
	RIGHT		Command	= "RIGHT"
	UP		Command	= "UP"
	D			= 'D'
	R			= 'R'
)

var (
	Origin	Coord	= Coord{0, 0}
	Steps		= map[Command]int{LEFT: 1, UP: 2, DOWN: 2, RIGHT: 3}
	Ranks		= [4][2]int{{1, -1}, {2, 0}, {2, 1}, {3, 2}}
)

type (
	Diver		struct{ Contestant }
	Hurdler		struct{ Contestant }
	Strategy	interface{ Apply(game Game) string }
	Diving		struct{ Race }
	Race		struct {
		gpu	string
		regs	[7]int
		players	[3]Player
	}
	Skating	struct{ Race }
	Command	string
	Archery	struct{ Race }
	Archer	struct{ Contestant }
	Game	interface {
		Place(Player) int
		Player(idx int) Player
		Update(gpu string, regs [7]int)
		Eval(cmd Command, playerIdx int) int
		isEOG() bool
	}
	Hurdling	struct{ Race }
	Score		map[Medal]int
	Skater		struct{ Contestant }
	Coord		struct {
		x	int
		y	int
	}
	Engine	struct {
		teamTotal	[3]int
		playerIdx	int
		races		map[string]Game
	}
	Player	interface {
		Update(score Score)
		Score() int
	}
	Contestant	struct {
		regs	[2]*int
		score	Score
	}
	StrategyFunc	func(g Game) string
	Medal		int
)

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
func NewArcher() *Archer {
	return &Archer{}
}
func (a Archer) coord() Coord {
	return Coord{x: *a.regs[0], y: *a.regs[1]}
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
	max := player.combo() + 1
	min := -player.combo()
	return d.normalize(float64(score), float64(min), float64(max))
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
func (e Engine) Exec() Command {
	bestAction := LEFT
	maxBias := -1 << 31
	geomMean := geometricMean(e.total(e.playerIdx), len(e.races))
	for _, cmd := range []Command{UP, DOWN, LEFT, RIGHT} {
		totalBias := 0
		for key, game := range e.races {
			bias := game.Eval(cmd, e.playerIdx)
			playerScore := game.Player(e.playerIdx).Score()
			place := game.Place(game.Player(e.playerIdx))
			if float64(playerScore) <= geomMean {
				bias = int(float64(bias) * 5)
			} else {
				bias = int(float64(bias) * 0.01)
			}
			opponentScores := []int{}
			for i := 0; i < 3; i++ {
				if i != e.playerIdx {
					opponentScores = append(opponentScores, game.Player(i).Score())
				}
			}
			maxOpponentScore := max(opponentScores)
			if playerScore < maxOpponentScore-10 {
				bias = int(float64(bias) * 0.5)
			}
			if place == 1 {
				bias = int(float64(bias) * 3)
			} else if place == 2 {
				bias = int(float64(bias) * 1.5)
			} else if place == 3 {
				bias *= 5
			}
			fmt.Fprintf(os.Stderr, "GAME: %8s, ACTION: %5s, BIAS: %5d, PLAYER SCORE: %3d, PLACE: %d, GEOM MEAN: %.2f\n", key, cmd, bias, playerScore, place, geomMean)
			totalBias += bias
		}
		if totalBias > maxBias {
			maxBias = totalBias
			bestAction = cmd
		}
	}
	return bestAction
}
func max(slice []int) int {
	maxValue := slice[0]
	for _, value := range slice {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}
func NewEngine(playerIdx int, games ...Game) Engine {
	races := make(map[string]Game, len(games))
	for _, game := range games {
		switch game.(type) {
		case *Hurdling:
			races[HURDLING] = game
		case *Diving:
			races[DIVING] = game
		case *Skating:
			races[SKATING] = game
		case *Archery:
			races[ARCHERY] = game
		default:
			fmt.Fprintln(os.Stderr, "UNKNOWN TYPE OF GAME:", game)
		}
	}
	return Engine{races: races, playerIdx: playerIdx}
}
func (e Engine) total(idx int) int {
	return e.teamTotal[idx]
}
func (e Engine) ListenAndServe(scanner *bufio.Scanner) {
	nbPlayers := 3
	for {
		for i := 0; i < nbPlayers; i++ {
			scanner.Scan()
			scoreInfo := strings.Fields(scanner.Text())
			hurdlingScore := Score{GOLD: toInt(scoreInfo[1]), SILVER: toInt(scoreInfo[2]), BRONZE: toInt(scoreInfo[3])}
			archeryScore := Score{GOLD: toInt(scoreInfo[4]), SILVER: toInt(scoreInfo[5]), BRONZE: toInt(scoreInfo[6])}
			skatingScore := Score{GOLD: toInt(scoreInfo[7]), SILVER: toInt(scoreInfo[8]), BRONZE: toInt(scoreInfo[9])}
			divingScore := Score{GOLD: toInt(scoreInfo[10]), SILVER: toInt(scoreInfo[11]), BRONZE: toInt(scoreInfo[12])}
			e.teamTotal[i] = toInt(scoreInfo[0])
			UpdatePlayer(e.races[HURDLING].Player(i), hurdlingScore)
			UpdatePlayer(e.races[ARCHERY].Player(i), archeryScore)
			UpdatePlayer(e.races[SKATING].Player(i), skatingScore)
			UpdatePlayer(e.races[DIVING].Player(i), divingScore)
		}
		gpu, regs := ParseState(scanner)
		UpdateGame(e.races[HURDLING], gpu, regs)
		gpu, regs = ParseState(scanner)
		UpdateGame(e.races[ARCHERY], gpu, regs)
		gpu, regs = ParseState(scanner)
		UpdateGame(e.races[SKATING], gpu, regs)
		gpu, regs = ParseState(scanner)
		UpdateGame(e.races[DIVING], gpu, regs)
		action := e.Exec()
		fmt.Println(action)
	}
}
func geometricMean(totalScore int, numGames int) float64 {
	if totalScore <= 0 {
		return 0
	}
	return math.Pow(float64(totalScore), 1.0/float64(numGames))
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
func NewHurdler() *Hurdler {
	return &Hurdler{}
}
func (h Hurdler) pos() int {
	return *h.regs[0]
}
func (h Hurdler) stuns() int {
	return *h.regs[1]
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
func (c Contestant) Score() int {
	return c.score.Calc()
}
func (c *Contestant) Update(score Score) {
	c.score = score
}
func (s Score) Calc() int {
	return s[GOLD]*3 + s[SILVER]
}
func NewScore(gold, silver, bronze int) Score {
	return Score{GOLD: gold, SILVER: silver, BRONZE: bronze}
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
func (s Skating) turnsLeft() int {
	return s.regs[6]
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
			oppPositions := []int{(p.spaces() + 1) % 10, (p.spaces() + 2) % 10, (p.spaces() + 2) % 10, (p.spaces() + 3) % 10}
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
func (sf StrategyFunc) Apply(game Game) string {
	return sf(game)
}
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
	fmt.Sscan(scanner.Text(), &gpu, &regs[0], &regs[1], &regs[2], &regs[3], &regs[4], &regs[5], &regs[6])
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
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)
	var playerIdx int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &playerIdx)
	var nbGames int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &nbGames)
	engine := NewEngine(playerIdx, NewHurdling(NewHurdler(), NewHurdler(), NewHurdler()), NewArchery(NewArcher(), NewArcher(), NewArcher()), NewSkating(NewSkater(), NewSkater(), NewSkater()), NewDiving(NewDiver(), NewDiver(), NewDiver()))
	engine.ListenAndServe(scanner)
}
