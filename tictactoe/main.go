package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

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

		// Output the chosen move in terms of the global board
		fmt.Println(bm.Row, bm.Col)

		game.Exec(PLAYER, Move{Row: bm.Row, Col: bm.Col})
	}
}
