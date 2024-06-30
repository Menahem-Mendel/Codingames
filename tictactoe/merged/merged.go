package main

import (
	"context"
	"time"
	"math"
	"math/rand"
	"sync"
	"fmt"
	"os"
)

const (
	SIZE			= 3
	EMPTY		Player	= 0
	OPPONENT	Player	= 1
	PLAYER		Player	= 2
)

type (
	Board	struct {
		board	[][]int
		player	Player
		size	int
	}
	Game	struct {
		games	[][]State
		player	Player
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
	Move	struct {
		Row	int
		Col	int
	}
)

func (b *Board) Player() Player {
	return b.player
}
func (b *Board) checkWin(player int) bool {
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
	for row := 0; row < b.size; row++ {
		for col := 0; col < b.size; col++ {
			if b.board[row][col] == int(EMPTY) {
				return false
			}
		}
	}
	return true
}
func (m Move) Apply(s State) State {
	game := s.(*Board)
	game.Exec(game.player, m)
	return game
}
func NewBoard(size int) *Board {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
		for j := range board[i] {
			board[i][j] = int(EMPTY)
		}
	}
	return &Board{board: board, player: PLAYER, size: size}
}
func (b *Board) Clone() State {
	newBoard := make([][]int, b.size)
	for i := range newBoard {
		newBoard[i] = make([]int, b.size)
		copy(newBoard[i], b.board[i])
	}
	return &Board{board: newBoard, player: b.player, size: b.size}
}
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
func (b *Board) Exec(p Player, action Action) {
	if action == nil {
		return
	}
	move := action.(Move)
	b.board[move.Row%b.size][move.Col%b.size] = int(b.player)
	b.player = 3 - b.player
}
func (b *Board) IsEOG() bool {
	return b.checkWin(int(PLAYER)) || b.checkWin(int(OPPONENT)) || b.isBoardFull()
}
func (b *Board) Eval(player Player) Result {
	if b.checkWin(int(player)) {
		return 1.0
	}
	if b.checkWin(3 - int(player)) {
		return -1.0
	}
	controlScore := 0.0
	if b.board[1][1] == int(player) {
		controlScore += 0.5
	}
	if b.board[0][0] == int(player) || b.board[0][2] == int(player) || b.board[2][0] == int(player) || b.board[2][2] == int(player) {
		controlScore += 0.25
	}
	return Result(controlScore)
}
func (g *Game) IsEOG() bool {
	return g.checkWin(int(PLAYER)) || g.checkWin(int(OPPONENT)) || g.isGamesFull()
}
func (g *Game) Eval(player Player) Result {
	var eval Result
	for _, col := range g.games {
		for _, game := range col {
			eval += game.Eval(player)
		}
	}
	return Result(eval)
}
func (g *Game) Player() Player {
	return g.player
}
func (g *Game) checkWin(player int) bool {
	win := true
	i, j := 0, 0
	for i < len(g.games) && j < len(g.games) {
		if g.games[i][j].Eval(Player(player)) != 1 || g.games[j][i].Eval(Player(player)) != 1 {
			win = false
		}
		i++
		j++
	}
	for _, row := range g.games {
		win = true
		for _, game := range row {
			if game.Eval(Player(player)) != 1 {
				win = false
			}
		}
	}
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
func NewGame(gameSize int, games ...State) *Game {
	gg := make([][]State, gameSize)
	for i := range gg {
		gg[i] = make([]State, gameSize)
		for j := range gg[i] {
			gg[i][j] = games[j]
		}
	}
	return &Game{games: gg}
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
func (g *Game) Clone() State {
	gg := make([][]State, len(g.games))
	for i := range gg {
		gg[i] = make([]State, len(g.games))
		copy(gg[i], g.games[i])
	}
	return &Game{games: gg}
}
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
func (g *Game) Exec(p Player, action Action) {
	if action == nil {
		fmt.Fprintln(os.Stderr, "ERROR: Received nil action")
		return
	}
	move := action.(Move)
	if move.Col < 0 || move.Row < 0 || move.Col >= len(g.games)*SIZE || move.Row >= len(g.games)*SIZE {
		fmt.Fprintf(os.Stderr, "ERROR: Game move out of range: [%d, %d]\n", move.Col, move.Row)
		return
	}
	subGameRow := move.Row / SIZE
	subGameCol := move.Col / SIZE
	subMoveRow := move.Row % SIZE
	subMoveCol := move.Col % SIZE
	if subGameRow >= len(g.games) || subGameCol >= len(g.games) {
		fmt.Fprintf(os.Stderr, "ERROR: Sub-game indices out of range: [%d, %d]\n", subGameRow, subGameCol)
		return
	}
	subGame := g.games[subGameRow][subGameCol]
	if subGame == nil {
		fmt.Fprintf(os.Stderr, "ERROR: Sub-game at indices [%d, %d] is nil\n", subGameRow, subGameCol)
		return
	} else if subGame.IsEOG() {
		fmt.Fprintf(os.Stderr, "ERROR: Sub-game has zero cells left \n")
		return
	}
	subMove := Move{Row: subMoveRow, Col: subMoveCol}
	if !isValidMove(subGame.Actions(), subMove) {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid move [%d, %d] in sub-game [%d, %d]\n", subMove.Row, subMove.Col, subGameRow, subGameCol)
		return
	}
	subGame.Exec(p, subMove)
	g.player = 3 - g.player
}
func treePolicy(ctx context.Context, node *Node) *Node {
	for !node.state.IsEOG() {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !node.isExpanded() {
				return expand(node)
			}
			node = bestChild(node, 1.0)
			if node == nil {
				fmt.Fprintln(os.Stderr, "ERROR: bestChild returned nil during treePolicy")
				return nil
			}
		}
	}
	return node
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
	stateClone := state.Clone()
	for !stateClone.IsEOG() {
		select {
		case <-ctx.Done():
			return -math.MaxFloat64
		default:
			actions := stateClone.Actions()
			var bestAction Action
			for _, action := range actions {
				move := action.(Move)
				if (move.Row == 1 && move.Col == 1) || (move.Row == 0 && move.Col == 0) || (move.Row == 0 && move.Col == 2) || (move.Row == 2 && move.Col == 0) || (move.Row == 2 && move.Col == 2) {
					bestAction = action
					break
				}
			}
			if bestAction == nil {
				bestAction = actions[rand.Intn(len(actions))]
			}
			stateClone.Exec(stateClone.Player(), bestAction)
		}
	}
	return float64(stateClone.Eval(stateClone.Player()))
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
						fmt.Fprintln(os.Stderr, "ERROR: treePolicy returned nil")
						return
					}
					reward := defaultPolicy(ctx, node.state)
					if reward == -math.MaxFloat64 {
						fmt.Fprintln(os.Stderr, "ERROR: defaultPolicy returned -Inf reward")
						return
					}
					backpropagate(ctx, node, reward)
				}
			}
		}()
	}
	wg.Wait()
	best := bestChild(root, 0)
	if best == nil {
		fmt.Fprintln(os.Stderr, "ERROR: bestChild returned nil")
		return nil
	}
	return best.action
}
func main() {
	const boardSize = 3
	games := make([]State, boardSize*boardSize)
	for i := range games {
		games[i] = NewBoard(boardSize)
	}
	game := NewGame(boardSize, games...)
	for {
		var opponentRow, opponentCol int
		fmt.Scan(&opponentRow, &opponentCol)
		game.Exec(OPPONENT, Move{Row: opponentRow, Col: opponentCol})
		var validActionCount int
		fmt.Scan(&validActionCount)
		var validMoves []Move
		for i := 0; i < validActionCount; i++ {
			var row, col int
			fmt.Scan(&row, &col)
			validMoves = append(validMoves, Move{Row: row, Col: col})
		}
		ctx, _ := context.WithTimeout(context.Background(), 90*time.Millisecond)
		if len(validMoves) == 0 {
			fmt.Fprintln(os.Stderr, "ERROR: No valid moves available")
			continue
		}
		bestMove := MCTS(ctx, game)
		if bestMove == nil {
			fmt.Fprintln(os.Stderr, "ERROR: MCTS didn't calculate the moves")
			bestMove = validMoves[0]
		}
		bm := bestMove.(Move)
		fmt.Println(bm.Row, bm.Col)
		game.Exec(PLAYER, Move{Row: bm.Row, Col: bm.Col})
	}
}
