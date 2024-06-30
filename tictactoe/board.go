package main

const SIZE = 3

type Move struct {
	Row int
	Col int
}

type Board struct {
	board  [][]int
	player Player
	size   int
}

func NewBoard(size int) *Board {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)

		for j := range board[i] {
			board[i][j] = int(EMPTY)
		}
	}
	return &Board{
		board:  board,
		player: PLAYER,
		size:   size,
	}
}

// Clone creates a deep copy of the game state
func (b *Board) Clone() State {
	newBoard := make([][]int, b.size)
	for i := range newBoard {
		newBoard[i] = make([]int, b.size)
		copy(newBoard[i], b.board[i])
	}
	return &Board{
		board:  newBoard,
		player: b.player,
		size:   b.size,
	}
}

// Actions returns a list of possible moves from the current state
func (b *Board) Actions() []Action {
	var actions []Action
	for row := 0; row < b.size; row++ {
		for col := 0; col < b.size; col++ {
			if b.board[row][col] != int(EMPTY) {
				continue
			}

			actions = append(actions, Move{Row: row, Col: col})
		}
	}

	return actions
}

// Exec applies a move to the game state
func (b *Board) Exec(p Player, action Action) {
	if action == nil {
		return
	}

	move := action.(Move)
	b.board[move.Row%b.size][move.Col%b.size] = int(b.player)
	b.player = 3 - b.player
}

// IsEOG checks if the game is over
func (b *Board) IsEOG() bool {
	return b.checkWin(int(PLAYER)) || b.checkWin(int(OPPONENT)) || b.isBoardFull()
}

// Eval evaluates the game state and returns the result from the perspective of the given player
func (b *Board) Eval(player Player) Result {
	// Example: Add weights to different winning scenarios or positions
	if b.checkWin(int(player)) {
		return 1.0
	}
	if b.checkWin(3 - int(player)) {
		return -1.0
	}
	// Evaluate based on control of center, corners, etc.
	controlScore := 0.0
	if b.board[1][1] == int(player) {
		controlScore += 0.5 // Center control
	}
	if b.board[0][0] == int(player) || b.board[0][2] == int(player) ||
		b.board[2][0] == int(player) || b.board[2][2] == int(player) {
		controlScore += 0.25 // Corner control
	}
	return Result(controlScore)
}

// Player returns the current player
func (b *Board) Player() Player {
	return b.player
}

// Helper methods

func (b *Board) checkWin(player int) bool {
	// Check rows
	for row := 0; row < b.size; row++ {
		win := true
		for col := 0; col < b.size; col++ {
			if b.board[row][col] != player {
				win = false
				break
			}
		}
		if win {
			return true
		}
	}
	// Check columns
	for col := 0; col < b.size; col++ {
		win := true
		for row := 0; row < b.size; row++ {
			if b.board[row][col] != player {
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
	for i := 0; i < b.size; i++ {
		if b.board[i][i] != player {
			win = false
			break
		}
	}
	if win {
		return true
	}
	win = true
	for i := 0; i < b.size; i++ {
		if b.board[i][b.size-1-i] != player {
			win = false
			break
		}
	}
	return win
}

func (b *Board) isBoardFull() bool {
	for _, row := range b.board {
		for _, cell := range row {
			if cell == int(EMPTY) {
				return false
			}
		}
	}
	return true
}

// Implementing Action interface for Move
func (m Move) Apply(s State) State {
	game := s.(*Board)
	game.Exec(game.player, m)
	return game
}
