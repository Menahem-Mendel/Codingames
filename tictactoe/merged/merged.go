package main

import (
	"math/rand"
	"os"
	"time"
	"math"
	"sync"
	"context"
	"fmt"
)

const (
	EMPTY		Player	= 0
	OPPONENT	Player	= 1
	PLAYER		Player	= 2
	SIZE			= 3
)

type (
	Move	struct {
		Row	int
		Col	int
	}
	Game	struct {
		board	[][]int
		player	Player
		size	int
	}
	Player	int
	State	interface {
		Clone() State
		Actions() []Action
		IsEOG() bool
		Player() Player
		Exec(p Player, a Action)
		Eval(p Player) Result
	}
	Result	float64
	Action	interface{ Apply(State) State }
	Node	struct {
		state		State
		parent		*Node
		children	[]*Node
		action		Action
		wins		float64
		visits		int
		sync.RWMutex
	}
)

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
func (m Move) Apply(s State) State {
	game := s.(*Game)
	game.Exec(game.player, m)
	return game
}
func NewGame(size int) *Game {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return &Game{board: board, player: PLAYER, size: size}
}
func (g *Game) Clone() State {
	newBoard := make([][]int, g.size)
	for i := range newBoard {
		newBoard[i] = make([]int, g.size)
		copy(newBoard[i], g.board[i])
	}
	return &Game{board: newBoard, player: g.player, size: g.size}
}
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
func (g *Game) Exec(p Player, action Action) {
	if action == nil {
		return
	}
	move := action.(Move)
	g.board[move.Row%g.size][move.Col%g.size] = int(g.player)
	g.player = 3 - g.player
}
func (g *Game) IsEOG() bool {
	return g.checkWin(int(PLAYER)) || g.checkWin(int(OPPONENT)) || g.isBoardFull()
}
func (g *Game) Eval(player Player) Result {
	if g.checkWin(int(player)) {
		return 1.0
	}
	if g.checkWin(3 - int(player)) {
		return -1.0
	}
	return 0.0
}
func (g *Game) Player() Player {
	return g.player
}
func (g *Game) checkWin(player int) bool {
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
func isValidMove(moves []Move, move Move) bool {
	r, c := index(moves, move)
	return r == -1 || c == -1
}
func index(moves []Move, move Move) (int, int) {
	for _, m := range moves {
		if (m.Col+1)%3 == move.Col+1 && (m.Row+1)%3 == move.Row+1 {
			return m.Row, m.Col
		}
	}
	return -1, -1
}
func expand(node *Node) *Node {
	if node.isExpanded() {
		return node.children[0]
	}
	actions := node.state.Actions()
	for _, action := range actions {
		childState := node.state.Clone()
		childState.Exec(node.state.Player(), action)
		childNode := NewNode(childState, node, action)
		node.children = append(node.children, childNode)
	}
	if len(node.children) <= 0 {
		return nil
	}
	return node.children[0]
}
func defaultPolicy(ctx context.Context, state State) float64 {
	clone := state.Clone()
	for !clone.IsEOG() {
		select {
		case <-ctx.Done():
			return -math.MaxFloat64
		default:
			actions := clone.Actions()
			action := actions[rand.Intn(len(actions))]
			clone.Exec(clone.Player(), action)
		}
	}
	return float64(clone.Eval(clone.Player()))
}
func backpropagate(ctx context.Context, node *Node, reward float64) {
	select {
	case <-ctx.Done():
		return
	default:
		if node == nil {
			return
		}
		node.Lock()
		node.visits++
		node.wins += reward
		node.Unlock()
		node = node.parent
	}
}
func bestChild(node *Node, c float64) *Node {
	max := math.Inf(-1)
	var nodes []*Node
	for _, child := range node.children {
		uctValue := (child.wins / float64(child.visits)) + c*math.Sqrt(math.Log(float64(node.visits))/float64(child.visits))
		if uctValue < max {
			continue
		}
		if uctValue == max {
			nodes = append(nodes, child)
			continue
		}
		max = uctValue
		nodes = []*Node{child}
	}
	if len(nodes) <= 0 {
		return node
	}
	return nodes[rand.Intn(len(nodes))]
}
func NewNode(state State, parent *Node, action Action) *Node {
	return &Node{state: state, parent: parent, action: action, children: make([]*Node, 0)}
}
func (n *Node) isExpanded() bool {
	return len(n.children) == len(n.state.Actions())
}
func root(state State) *Node {
	return NewNode(state, nil, nil)
}
func MCTS(ctx context.Context, state State) Action {
	root := root(state)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					node := treePolicy(ctx, root)
					if node == nil {
						return
					}
					reward := defaultPolicy(ctx, node.state)
					if reward == -math.MaxFloat64 {
						return
					}
					backpropagate(ctx, node, reward)
				}
			}
		}()
	}
	wg.Wait()
	return bestChild(root, 0).action
}
func treePolicy(ctx context.Context, node *Node) *Node {
	node.Lock()
	defer node.Unlock()
	for !node.state.IsEOG() {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !node.isExpanded() {
				return expand(node)
			}
			node = bestChild(node, 1.0)
		}
	}
	return node
}
func main() {
	const boardSize = 3
	game := NewGame(boardSize)
	games := make([][]*Game, boardSize)
	for i := range games {
		games[i] = make([]*Game, boardSize)
		for j := range games[i] {
			games[i][j] = NewGame(boardSize)
		}
	}
	for {
		var opponentRow, opponentCol int
		fmt.Scan(&opponentRow, &opponentCol)
		if opponentRow != -1 && opponentCol != -1 {
			subBoardRow := opponentRow / boardSize
			subBoardCol := opponentCol / boardSize
			cellRow := opponentRow % boardSize
			cellCol := opponentCol % boardSize
			game = games[subBoardRow][subBoardCol]
			game.Exec(OPPONENT, Move{Row: cellRow, Col: cellCol})
		}
		var validActionCount int
		fmt.Scan(&validActionCount)
		var validMoves []Move
		for i := 0; i < validActionCount; i++ {
			var row, col int
			fmt.Scan(&row, &col)
			validMoves = append(validMoves, Move{Row: row, Col: col})
		}
		ctx, _ := context.WithTimeout(context.Background(), 80*time.Millisecond)
		if len(validMoves) == 0 {
			fmt.Fprintln(os.Stderr, "ERROR: No valid moves available")
			continue
		}
		row, col := validMoves[rand.Intn(len(validMoves))].Row, validMoves[rand.Intn(len(validMoves))].Col
		subBoardRow := row / boardSize
		subBoardCol := col / boardSize
		game = games[subBoardRow][subBoardCol]
		bestMove := MCTS(ctx, game)
		if bestMove == nil {
			fmt.Fprintln(os.Stderr, "ERROR: MCTS didn't calculate the moves")
			fmt.Println(row, col)
			continue
		}
		bm := bestMove.(Move)
		fmt.Fprintf(os.Stderr, "move row %2d, col %2d\n", bm.Row, bm.Col)
		fmt.Fprintf(os.Stderr, "vald row %2d, col %2d\n", row, col)
		game.Exec(PLAYER, bm)
		globalRow := subBoardRow*boardSize + bm.Row
		globalCol := subBoardCol*boardSize + bm.Col
		fmt.Fprintf(os.Stderr, "glob row %2d, col %2d\n", globalRow, globalCol)
		if !isValidMove(validMoves, Move{Row: globalRow, Col: globalCol}) {
			fmt.Fprintln(os.Stderr, "ERROR: MCTS selected an invalid move")
			bm = validMoves[0]
		}
		fmt.Println(globalRow, globalCol)
	}
}
