package main

import (
	"fmt"
	"os"
)

type Game struct {
	games  [][]State
	player Player
}

func NewGame(gameSize int, games ...State) *Game {
	gg := make([][]State, gameSize)
	for i := range gg {
		gg[i] = make([]State, gameSize)

		for j := range gg[i] {
			gg[i][j] = games[j]
		}
	}

	return &Game{
		games: gg,
	}
}

// Clone creates a deep copy of the game state
func (g *Game) Clone() State {
	gg := make([][]State, len(g.games))
	for i := range gg {
		gg[i] = make([]State, len(g.games))
		copy(gg[i], g.games[i])
	}

	return &Game{
		games: gg,
	}
}

// Actions returns a list of possible moves from the current state
func (g *Game) Actions() []Action {
	var actions []Action
	for i, col := range g.games {
		for j, game := range col {
			for _, action := range game.Actions() {
				move := action.(Move)

				move.Row = move.Row + j*len(g.games)
				move.Col = move.Col + i*len(g.games)

				actions = append(actions, move)
			}
		}
	}

	return actions
}

// Exec applies a move to the game state
func (g *Game) Exec(p Player, action Action) {
	if action == nil {
		fmt.Fprintln(os.Stderr, "ERROR: Received nil action")
		return
	}

	move := action.(Move)

	// Ensure move is within the valid range for the main game
	if move.Col < 0 || move.Row < 0 || move.Col >= len(g.games)*SIZE || move.Row >= len(g.games)*SIZE {
		fmt.Fprintf(os.Stderr, "ERROR: Game move out of range: [%d, %d]\n", move.Col, move.Row)
		return
	}

	// Calculate the sub-game indices
	subGameRow := move.Row / SIZE
	subGameCol := move.Col / SIZE

	// Calculate the move within the sub-game
	subMoveRow := move.Row % SIZE
	subMoveCol := move.Col % SIZE

	// Ensure the sub-game indices are within bounds
	if subGameRow >= len(g.games) || subGameCol >= len(g.games) {
		fmt.Fprintf(os.Stderr, "ERROR: Sub-game indices out of range: [%d, %d]\n", subGameRow, subGameCol)
		return
	}

	// Get the sub-game
	subGame := g.games[subGameRow][subGameCol]
	if subGame == nil {
		fmt.Fprintf(os.Stderr, "ERROR: Sub-game at indices [%d, %d] is nil\n", subGameRow, subGameCol)
		return
	} else if subGame.IsEOG() {
		fmt.Fprintf(os.Stderr, "ERROR: Sub-game has zero cells left \n")
		return
	}

	// Create the move for the sub-game
	subMove := Move{Row: subMoveRow, Col: subMoveCol}

	// Ensure the move is valid within the sub-game
	if !isValidMove(subGame.Actions(), subMove) {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid move [%d, %d] in sub-game [%d, %d]\n", subMove.Row, subMove.Col, subGameRow, subGameCol)
		return
	}

	// Apply the move to the sub-game
	subGame.Exec(p, subMove)

	// Switch the player
	g.player = 3 - g.player
}

// IsEOG checks if the game is over
func (g *Game) IsEOG() bool {

	return g.checkWin(int(PLAYER)) || g.checkWin(int(OPPONENT)) || g.isGamesFull()
}

// Eval evaluates the game state and returns the result from the perspective of the given player
func (g *Game) Eval(player Player) Result {
	var eval Result
	for _, col := range g.games {
		for _, game := range col {
			eval += game.Eval(player)
		}
	}
	return Result(eval)
}

// Player returns the current player
func (g *Game) Player() Player {
	return g.player
}

// Helper methods

func (g *Game) checkWin(player int) bool {
	// Check diagonals
	win := true
	i, j := 0, 0
	for i < len(g.games) && j < len(g.games) {
		if g.games[i][j].Eval(Player(player)) != 1 || g.games[j][i].Eval(Player(player)) != 1 {
			win = false
		}

		i++
		j++
	}

	// Check rows
	for _, row := range g.games {
		win = true
		for _, game := range row {
			if game.Eval(Player(player)) != 1 {
				win = false
			}
		}

	}

	// Check columns
	for row, r := range g.games {
		win = true
		for col := range r {
			if g.games[col][row].Eval(Player(player)) != 1 {
				win = false
			}
		}
		if win {
			return win
		}
	}
	return win
}

func (g *Game) isGamesFull() bool {
	for _, row := range g.games {
		for _, game := range row {
			if !game.IsEOG() {
				return false
			}
		}
	}
	return true
}

func isValidMove(actions []Action, move Move) bool {
	for _, action := range actions {
		m := action.(Move)
		if m.Row == move.Row && m.Col == move.Col {
			return true
		}
	}
	return false
}
