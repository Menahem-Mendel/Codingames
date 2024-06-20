package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
)

// Constants representing different elements and actions in the games
const (
	DOT    = '.'         // Represents an empty space on the track in Hurdle Race
	HURDLE = '#'         // Represents a hurdle on the track in Hurdle Race
	L      = 'l'         // Represents a left move in Roller Skating
	D      = 'D'         // Represents a down move in Roller Skating
	R      = 'R'         // Represents a right move in Roller Skating
	U      = 'U'         // Represents an up move in Roller Skating
	EOG    = "GAME_OVER" // Represents the end of the game
)

// Command represents possible player actions
type Command string

// Constants for possible player actions
const (
	LEFT  Command = "LEFT"
	DOWN  Command = "DOWN"
	RIGHT Command = "RIGHT"
	UP    Command = "UP"
)

func main() {
	// Scanner for reading input from stdin
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	var engine Engine

	// Read player index (0, 1, or 2)
	var playerIdx int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &playerIdx)

	// Read the number of games being played
	var gamesCount int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &gamesCount)

	// Initialize the games within the engine
	engine.Races = make([]Game, gamesCount)
	engine.Races[0] = NewHurdleRace(playerIdx)
	engine.Races[1] = NewArchery(playerIdx)
	engine.Races[2] = NewRollerSkating(playerIdx)
	engine.Races[3] = NewDiving(playerIdx)

	for {
		for i := 0; i < 3; i++ {
			scanner.Scan()
			scoreInfo := scanner.Text()
			fmt.Fprintf(os.Stderr, "SCORE[%d] %s\n", i, scoreInfo)
		}

		// Read the state for each game and update the engine
		for i := 0; i < gamesCount; i++ {
			var gpu string
			regs := make([]int, 7)
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &gpu,
				&regs[0], &regs[1], &regs[2], &regs[3], &regs[4], &regs[5], &regs[6])

			fmt.Fprintf(os.Stderr, "GAME[%d] %s - %v\n", i, gpu, regs)
			// Update the state of the game with the GPU string and register values
			update(engine.Races[i], gpu, regs)
		}

		// Determine the best action to take based on the current game states
		action := engine.Exec(nil)

		// Output the chosen action
		fmt.Println(action)
	}
}

type State interface {
	// Update updates the state of the game using the GPU string and register values.
	// This method is called every turn to refresh the game's internal state.
	Update(gpu string, regs []int)
}

// Game interface represents the common behavior of all mini-games.
// Each game must implement methods to update its state and evaluate actions.
type Game interface {
	State

	Place() int

	Normalize(score, min, max int) int

	// Eval evaluates the score for a given action in the context of the game.
	// It returns an integer score representing the benefit of performing the action.
	Eval(c Command) int

	// IsEOG checks if the game has ended.
	IsEOG() bool
}

// Engine manages multiple games and determines the best action to take.
// It evaluates the scores of possible actions across all games and selects the optimal one.
type Engine struct {
	Races []Game // List of games managed by the engine
}

// Exec returns the best action based on the current states of all games.
// It iterates over possible actions, evaluates them, and selects the one with the highest total score.
func (e Engine) Exec(s StrategyFunc) Command {
	bestAction := LEFT
	bestScore := -1 << 31 // Initialize to the minimum possible integer value

	// Evaluate each possible action
	for _, action := range []Command{UP, DOWN, LEFT, RIGHT} {
		totalScore := 0

		// Sum the scores for each game for the current action
		for i, game := range e.Races {
			score := game.Eval(action)
			normalized := game.Normalize(score, -10, 10)

			switch game.Place() {
			case 2:
				normalized *= 2
			case 3:
				normalized *= 5
			}

			fmt.Fprintf(os.Stderr, "NORMALIZED: PLACE - %d, RACE[%d] %s, %d -> %d\n", game.Place(), i, action, score, normalized)
			totalScore += normalized

		}

		// Choose the action with the highest total score
		if totalScore > bestScore {
			bestScore = totalScore
			bestAction = action
		}
	}

	return bestAction
}

// Strategy interface represents a strategy for determining actions.
// It allows different strategies to be implemented and applied to games.
type Strategy interface {
	// Apply the strategy to the game and return the chosen action.
	Apply(game Game) string
}

// StrategyFunc is a function type that implements the Strategy interface.
// It allows a function to be used as a strategy by implementing the Apply method.
type StrategyFunc func(g Game) string

// Apply executes the strategy function on the game and returns the chosen action.
func (sf StrategyFunc) Apply(game Game) string {
	return sf(game)
}

// HurdleRace represents the state and behavior of the Hurdle Race mini-game.
// It maintains the track layout, player positions, and game status.
type HurdleRace struct {
	playerIdx int
	isEOG     bool
	track     string              // Track layout with hurdles
	players   []*HurdleRacePlayer // List of players in the race
}

// NewHurdleRace initializes a new HurdleRace game instance.
func NewHurdleRace(playerIdx int) *HurdleRace {
	players := make([]*HurdleRacePlayer, 0, 3)

	for i := 0; i < cap(players); i++ {
		players = append(players, &HurdleRacePlayer{})
	}

	return &HurdleRace{
		players:   players,
		playerIdx: playerIdx,
	}
}

// Update updates the state of the Hurdle Race game using the GPU string and register values.
func (hr *HurdleRace) Update(gpu string, regs []int) {
	if gpu == EOG {
		hr.isEOG = true
		return
	}

	hr.players[0].Update(regs[0], regs[3])
	hr.players[1].Update(regs[1], regs[4])
	hr.players[2].Update(regs[2], regs[5])

	hr.track = gpu
}

// Eval evaluates the score for a given action in the Hurdle Race game.
func (hr HurdleRace) Eval(c Command) int {
	score := 0

	if hr.IsStunned() {
		return score
	}

	if hr.IsStunMove(c, hr.players[hr.playerIdx].pos) {
		score -= 3
	}

	dist := len(hr.track) - hr.players[hr.playerIdx].pos

	switch c {
	case UP:
		score += calcScore(dist, 2)
	case LEFT:
		score += calcScore(dist, 1)
	case DOWN:
		score += calcScore(dist, 2)
	case RIGHT:
		score += calcScore(dist, 3)
	}

	return score
}

func (hr HurdleRace) IsStunned() bool {
	return hr.players[hr.playerIdx].stuns > 0
}

func (hr *HurdleRace) IsStunMove(c Command, pos int) bool {
	if len(hr.track) < pos {
		return false
	}

	track := hr.track[pos:]

	switch c {
	case UP:
		return checkForHurdle(track, 2)
	case LEFT:
		return checkForHurdle(track, 1)
	case DOWN:
		return checkForHurdle(track, 1) || checkForHurdle(track, 2)
	case RIGHT:
		return checkForHurdle(track, 1) || checkForHurdle(track, 2) || checkForHurdle(track, 3)
	}

	return false
}

// HurdleRace Normalize Method
func (hr HurdleRace) Normalize(score, min, max int) int {
	// The actual min and max for HurdleRace are [-2, 3]
	actualMin, actualMax := -2, 3
	return normalize(score, actualMin, actualMax, min, max)
}

func (hr HurdleRace) Place() int {
	place := 1
	player := hr.players[hr.playerIdx]
	for _, opp := range hr.players {
		if player == opp || player.pos-player.stuns > opp.pos-player.stuns {
			continue
		}
		place++
	}

	return place
}

// IsEOG checks if the Hurdle Race game has ended.
func (hr HurdleRace) IsEOG() bool {
	return hr.isEOG
}

// HurdleRacePlayer represents a player in the Hurdle Race mini-game.
// It maintains the player's position on the track and the number of stuns.
type HurdleRacePlayer struct {
	pos   int // Current position on the track
	stuns int // Number of stuns
}

// Update updates the state of a Hurdle Race player using the given position and stun count.
func (hrp *HurdleRacePlayer) Update(pos, stuns int) {
	hrp.pos, hrp.stuns = pos, stuns
}

// Archery represents the state and behavior of the Archery mini-game.
// It maintains player positions and the wind strengths affecting the game.
type Archery struct {
	isEOG     bool
	playerIdx int
	winds     []int
	players   []*ArcheryPlayer // List of players in the Archery game
}

// NewArchery initializes a new Archery game instance.
func NewArchery(playerIdx int) *Archery {
	players := make([]*ArcheryPlayer, 0, 3)

	for i := 0; i < cap(players); i++ {
		players = append(players, &ArcheryPlayer{})
	}

	return &Archery{
		players:   players,
		playerIdx: playerIdx,
	}
}

// Update updates the state of the Archery game using the GPU string and register values.
func (a *Archery) Update(winds string, regs []int) {
	if winds == EOG {
		a.isEOG = true
		return
	}

	a.players[0].Update(regs[0], regs[1])
	a.players[1].Update(regs[2], regs[3])
	a.players[2].Update(regs[4], regs[5])

	a.winds = make([]int, 0, 15)
	for _, r := range winds {
		n, err := strconv.Atoi(string(r))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		a.winds = append(a.winds, n)
	}
}

// Eval evaluates the score for a given action in the Archery game.
func (a Archery) Eval(c Command) int {
	score := 0
	wind := a.winds[0]

	coord := a.players[a.playerIdx].coord
	origin := Coord{}

	switch c {
	case LEFT:
		coord.x -= wind
	case UP:
		coord.y -= wind
	case RIGHT:
		coord.x += wind
	case DOWN:
		coord.y += wind
	}

	// Ensure the new coordinates are within the bounds [-20, 20]
	coord.x = clamp(coord.x, -20, 20)
	coord.y = clamp(coord.y, -20, 20)

	deltaPrev := dist(a.players[a.playerIdx].coord, origin)
	delta := dist(coord, origin)

	// Calculate the change in distance and normalize it to the range [3, -2]
	score = deltaPrev - delta

	return score
}

// Archery Normalize Method
func (a Archery) Normalize(score, min, max int) int {
	// The actual min and max for Archery are [-9, 9]
	actualMin, actualMax := -9, 9
	return normalize(score, actualMin, actualMax, min, max)
}
func (a Archery) Place() int {
	place := 1
	origin := Coord{0, 0}
	player := a.players[a.playerIdx]
	for _, opp := range a.players {
		if player == opp || dist(player.coord, origin) < dist(opp.coord, origin) {
			continue
		}
		place++
	}

	return place
}

// IsEOG checks if the Archery game has ended.
func (a Archery) IsEOG() bool {
	return a.isEOG
}

// Coord represents coordinates with x and y values.
type Coord struct {
	x int // x-coordinate
	y int // y-coordinate
}

// ArcheryPlayer represents a player in the Archery mini-game.
// It maintains the player's x and y coordinates on the target.
type ArcheryPlayer struct {
	coord Coord // Coordinate of the player
}

// Update updates the state of an Archery player using the given x and y coordinates.
func (ap *ArcheryPlayer) Update(x, y int) {
	ap.coord.x, ap.coord.y = x, y
}

// Diving represents the state and behavior of the Diving mini-game.
// It manages the diving sequence and player scores.
type Diving struct {
	isEOG     bool
	playerIdx int
	goal      string // Current complete diving goal
	players   []*DivingPlayer
}

// NewDiving initializes a new Diving game instance.
func NewDiving(playerIdx int) *Diving {
	players := make([]*DivingPlayer, 0, 3)

	for i := 0; i < cap(players); i++ {
		players = append(players, &DivingPlayer{})
	}

	return &Diving{
		players:   players,
		playerIdx: playerIdx,
	}
}

// Update updates the state of the Diving game using the GPU string and register values.
func (d *Diving) Update(gpu string, regs []int) {
	if gpu == EOG {
		d.isEOG = true
		return
	}
	d.players[0].Update(regs[0], regs[3])
	d.players[1].Update(regs[1], regs[4])
	d.players[2].Update(regs[2], regs[5])

	d.goal = gpu
}

// Eval evaluates the score for a given action in the Diving game.
func (d Diving) Eval(c Command) int {
	score := 0
	player := d.players[d.playerIdx]

	// Evaluate score based on current player's combo and possible future score
	// The more the combo, the greater the loss if the next step breaks it
	if len(d.goal) > 0 && rune(d.goal[0]) != rune(c[0]) {
		return -player.combo

	}

	// Simulate matching the remaining sequence
	for i := range d.goal {
		score += player.combo + i
	}

	return score
}

func (d Diving) Normalize(score, min, max int) int {
	// The actual maximum score is based on the remaining sequence and combos that can be achieved from now on
	// Calculate the maximum possible score from the remaining sequence
	actualMax := 0
	player := d.players[d.playerIdx]
	for i := range d.goal {
		actualMax += player.combo + i
	}

	return normalize(score, -d.players[d.playerIdx].combo, actualMax, min, max)
}

func (d Diving) Place() int {
	place := 1
	player := d.players[d.playerIdx]
	for _, opp := range d.players {
		if player == opp || player.score > opp.score+opp.combo+1 {
			continue
		}
		place++
	}
	return place
}

// IsEOG checks if the Diving game has ended.
func (d Diving) IsEOG() bool {
	return d.isEOG
}

// DivingPlayer represents a player in the Diving mini-game.
// It maintains the player's current score and combo multiplier.
type DivingPlayer struct {
	score int // Current score of the player
	combo int // Current combo multiplier of the player
}

// Update updates the state of a Diving player using the given score and combo multiplier.
func (dp *DivingPlayer) Update(score, combo int) {
	dp.score, dp.combo = score, combo
}

// Rank represents the ranking of a move in terms of movement and risk.
type Rank struct {
	move int // Number of spaces to move
	risk int // Risk level associated with the move
}

// RollerSkating represents the state and behavior of the Roller Skating mini-game.
// It manages player positions and risk levels during the race.
type RollerSkating struct {
	isEOG         bool
	playerIdx     int
	order         map[Command]Rank
	players       []*RollerSkatingPlayer
	turnsRemained int
}

// NewRollerSkating initializes a new Roller Skating game instance.
func NewRollerSkating(playerIdx int) *RollerSkating {
	players := make([]*RollerSkatingPlayer, 0, 3)

	for i := 0; i < cap(players); i++ {
		players = append(players, &RollerSkatingPlayer{})
	}

	return &RollerSkating{
		players:   players,
		playerIdx: playerIdx,
		order:     make(map[Command]Rank, 4),
	}
}

// Update updates the state of the Roller Skating game using the GPU string and register values.
func (rs *RollerSkating) Update(actions string, regs []int) {
	if actions == EOG {
		rs.isEOG = true
		return
	}

	rs.players[0].Update(regs[0], regs[3])
	rs.players[1].Update(regs[1], regs[4])
	rs.players[2].Update(regs[2], regs[5])

	for i, r := range actions {
		rank := Rank{0, 0}
		switch i {
		case 0:
			rank = Rank{
				move: 1,
				risk: -1,
			}
		case 1:
			rank = Rank{
				move: 2,
				risk: 0,
			}
		case 2:
			rank = Rank{
				move: 2,
				risk: 1,
			}
		case 3:
			rank = Rank{
				move: 3,
				risk: 2,
			}
		}

		switch r {
		case U:
			rs.order[UP] = rank
		case D:
			rs.order[DOWN] = rank
		case L:
			rs.order[LEFT] = rank
		case R:
			rs.order[RIGHT] = rank
		}
	}

	rs.turnsRemained = regs[6]
}

// Eval evaluates the score for a given action in the Roller Skating game.
func (rs RollerSkating) Eval(c Command) int {
	score := 0
	if rs.players[rs.playerIdx].risk < 0 {
		return score
	}

	rank := rs.order[c]
	player := rs.players[rs.playerIdx]

	if stunned := player.risk + rank.risk; stunned > 4 {
		score -= 3
		// if the risk is too high then it's worth the trouble
		// score += (3 % (stunned - 4)) / 2
		return score
	}

	risk := player.risk + rank.risk
	if risk > 3 {
		score -= 1
	}

	if risk < 1 {
		score += 1
	}

	// newPos := (player.spaces + rank.move) % 10
	// for i, otherPlayer := range rs.players[1:] {
	// 	if newPos != otherPlayer.spaces%10 {
	// 		continue
	// 	}

	// 	// Increase risk for both players
	// 	player.risk += 2
	// }

	score += rank.move

	return score
}

func (rs RollerSkating) Normalize(score, min, max int) int {
	actualMax := 3
	actualMin := -2

	return normalize(score, actualMin, actualMax, min, max)
}

func (rs RollerSkating) Place() int {
	place := 1
	player := rs.players[rs.playerIdx]
	stuns := 0
	if player.risk < 0 {
		stuns = player.risk
	}
	for _, opp := range rs.players {
		oppStuns := 0
		if opp.risk < 0 {
			oppStuns = opp.risk
		}
		if player == opp || player.spaces+stuns > opp.spaces+oppStuns {
			continue
		}
		place++
	}
	return place
}

// IsEOG checks if the Roller Skating game has ended.
func (rs RollerSkating) IsEOG() bool {
	return rs.isEOG
}

// RollerSkatingPlayer represents a player in the Roller Skating mini-game.
// It maintains the player's position on the track and current risk level.
type RollerSkatingPlayer struct {
	spaces int // Spaces travelled by player
	risk   int // Risk of player or stun timer as a negative number if stunned
}

// Update updates the state of a Roller Skating player using the given position and risk level.
func (rsp *RollerSkatingPlayer) Update(spaces, risk int) {
	rsp.spaces, rsp.risk = spaces, risk
}

func update(s State, gpu string, regs []int) {
	s.Update(gpu, regs)
}

// dist calculates the distance for a coordinate in the Archery game.
func dist(a, b Coord) int {
	return int(math.Sqrt(float64((a.x-b.x)*(a.x-b.x) + (a.y-b.y)*(a.y-b.y))))
}

func calcScore(remained, move int) int {
	if remained < move && remained > 0 {
		return remained
	}
	return move
}

func checkForHurdle(track string, distance int) bool {
	return len(track) > distance && track[distance] == HURDLE
}

// clamp ensures the value is within the given min and max bounds.
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Generic Normalize Function
func normalize(score, actualMin, actualMax, min, max int) int {
	// Scale score to the [min, max] range based on actualMin and actualMax
	normalized := int(float64((score-actualMin)*(max-min))/float64(actualMax-actualMin)) + min

	// Clamp the normalized score to the range [min, max]
	if normalized > max {
		return max
	}
	if normalized < min {
		return min
	}
	return normalized
}
