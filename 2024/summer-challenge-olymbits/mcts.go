package main

// import (
// 	"fmt"
// 	"math"
// 	"math/rand"
// 	"time"
// )

// // Define constants
// const (
// 	MaxFirstTurnTime = 1000 * time.Millisecond
// 	MaxTurnTime      = 50 * time.Millisecond
// )

// // TreeNode represents a node in the MCTS tree
// type TreeNode struct {
// 	state       *Engine
// 	parent      *TreeNode
// 	children    []*TreeNode
// 	visits      int
// 	totalReward float64
// 	action      Command
// }

// func NewTreeNode(state *Engine, parent *TreeNode, action Command) *TreeNode {
// 	return &TreeNode{
// 		state:       state,
// 		parent:      parent,
// 		action:      action,
// 		visits:      0,
// 		children:    nil,
// 		totalReward: 0.0,
// 	}
// }

// func (n *TreeNode) UCT() float64 {
// 	if n.visits == 0 {
// 		return math.Inf(1)
// 	}
// 	return n.totalReward/float64(n.visits) + math.Sqrt(2*math.Log(float64(n.parent.visits))/float64(n.visits))
// }

// func (n *TreeNode) SelectBestChild() *TreeNode {
// 	var bestChild *TreeNode
// 	bestValue := -math.Inf(1)
// 	for _, child := range n.children {
// 		uctValue := child.UCT()
// 		if uctValue > bestValue {
// 			bestValue = uctValue
// 			bestChild = child
// 		}
// 	}
// 	return bestChild
// }

// func (n *TreeNode) Expand() {
// 	commands := []Command{UP, DOWN, LEFT, RIGHT}
// 	for _, cmd := range commands {
// 		newState := n.state.Copy()
// 		newState.ApplyCommand(cmd)
// 		childNode := NewTreeNode(newState, n, cmd)
// 		n.children = append(n.children, childNode)
// 	}
// }

// func (n *TreeNode) Simulate() float64 {
// 	simulatedState := n.state.Copy()
// 	for !simulatedState.IsGameOver() {
// 		randomCmd := n.getRandomCommand()
// 		simulatedState.ApplyCommand(randomCmd)
// 	}
// 	return simulatedState.Evaluate()
// }

// func (n *TreeNode) Backpropagate(reward float64) {
// 	currentNode := n
// 	for currentNode != nil {
// 		currentNode.visits++
// 		currentNode.totalReward += reward
// 		currentNode = currentNode.parent
// 	}
// }

// func (n *TreeNode) getRandomCommand() Command {
// 	commands := []Command{UP, DOWN, LEFT, RIGHT}
// 	return commands[rand.Intn(len(commands))]
// }

// func MonteCarloTreeSearch(initialState *Engine, maxTime time.Duration) Command {
// 	root := NewTreeNode(initialState, nil, LEFT)
// 	simulations := 0
// 	results := make(chan bool)

// 	endTime := time.Now().Add(maxTime)
// 	go func() {
// 		for {
// 			if time.Now().After(endTime) {
// 				results <- true
// 				return
// 			}
// 			node := root
// 			for len(node.children) != 0 {
// 				node = node.SelectBestChild()
// 			}
// 			if node.visits > 0 {
// 				node.Expand()
// 				node = node.SelectBestChild()
// 			}
// 			reward := node.Simulate()
// 			node.Backpropagate(reward)
// 			simulations++
// 		}
// 	}()

// 	select {
// 	case <-results:
// 		// Timeout reached, stop MCTS
// 	}

// 	bestChild := root.SelectBestChild()
// 	fmt.Printf("Simulations: %d\n", simulations)
// 	return bestChild.action
// }
