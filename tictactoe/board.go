package main

const SIZE = 3

type Move struct {
	Row int
	Col int
}

type Game struct {
	board  [][]int
	player Player
	size   int
}

func NewGame(size int) *Game {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return &Game{
		board:  board,
		player: PLAYER,
		size:   size,
	}
}

// Clone creates a deep copy of the game state
func (g *Game) Clone() State {
	newBoard := make([][]int, g.size)
	for i := range newBoard {
		newBoard[i] = make([]int, g.size)
		copy(newBoard[i], g.board[i])
	}
	return &Game{
		board:  newBoard,
		player: g.player,
		size:   g.size,
	}
}

// Actions returns a list of possible moves from the current state
func (g *Game) Actions() []Action {
	var actions []Action
	for row := 0; row < g.size; row++ {
		for col := 0; col < g.size; col++ {
			if g.board[row][col] != int(EMPTY) {
				continue
			}

			actions = append(actions, Move{Row: row, Col: col})
		}
	}

	return actions
}

// Exec applies a move to the game state
func (g *Game) Exec(p Player, action Action) {
	if action == nil {
		return
	}

	move := action.(Move)
	g.board[move.Row%g.size][move.Col%g.size] = int(g.player)
	g.player = 3 - g.player
}

// IsEOG checks if the game is over
func (g *Game) IsEOG() bool {
	return g.checkWin(int(PLAYER)) || g.checkWin(int(OPPONENT)) || g.isBoardFull()
}

// Eval evaluates the game state and returns the result from the perspective of the given player
func (g *Game) Eval(player Player) Result {
	if g.checkWin(int(player)) {
		return 1.0
	}
	if g.checkWin(3 - int(player)) {
		return -1.0
	}
	return 0.0
}

// Player returns the current player
func (g *Game) Player() Player {
	return g.player
}

// Helper methods

func (g *Game) checkWin(player int) bool {
	// Check rows
	for row := 0; row < g.size; row++ {
		win := true
		for col := 0; col < g.size; col++ {
			if g.board[row][col] != player {
				win = false
				break
			}
		}
		if win {
			return true
		}
	}
	// Check columns
	for col := 0; col < g.size; col++ {
		win := true
		for row := 0; row < g.size; row++ {
			if g.board[row][col] != player {
				win = false
				break
			}
		}
		if win {
			return true
		}
	}
	// Check diagonals
	win := true
	for i := 0; i < g.size; i++ {
		if g.board[i][i] != player {
			win = false
			break
		}
	}
	if win {
		return true
	}
	win = true
	for i := 0; i < g.size; i++ {
		if g.board[i][g.size-1-i] != player {
			win = false
			break
		}
	}
	return win
}

func (g *Game) isBoardFull() bool {
	for row := 0; row < g.size; row++ {
		for col := 0; col < g.size; col++ {
			if g.board[row][col] == int(EMPTY) {
				return false
			}
		}
	}
	return true
}

// Implementing Action interface for Move
func (m Move) Apply(s State) State {
	game := s.(*Game)
	game.Exec(game.player, m)
	return game
}
