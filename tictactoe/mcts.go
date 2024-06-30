package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
)

type Player int

const (
	EMPTY    Player = 0
	OPPONENT Player = 1
	PLAYER   Player = 2
)

type State interface {
	Clone() State
	Actions() []Action
	IsEOG() bool
	Player() Player
	Exec(p Player, a Action)
	Eval(p Player) Result
}

type Result float64

type Action interface {
	Apply(State) State
}

type Node struct {
	state    State
	parent   *Node
	children []*Node
	action   Action
	wins     float64
	visits   int

	sync.RWMutex
}

func NewNode(state State, parent *Node, action Action) *Node {
	return &Node{
		state:    state,
		parent:   parent,
		action:   action,
		children: make([]*Node, 0),
	}
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

// defaultPolicy simulates a random playout from the given state
func defaultPolicy(ctx context.Context, state State) float64 {
	stateClone := state.Clone()
	for !stateClone.IsEOG() {
		select {
		case <-ctx.Done():
			return -math.MaxFloat64
		default:
			actions := stateClone.Actions()
			// Example: Prefer center and corners in rollouts
			var bestAction Action
			for _, action := range actions {
				move := action.(Move)
				if (move.Row == 1 && move.Col == 1) || // Center
					(move.Row == 0 && move.Col == 0) || (move.Row == 0 && move.Col == 2) ||
					(move.Row == 2 && move.Col == 0) || (move.Row == 2 && move.Col == 2) { // Corners
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
