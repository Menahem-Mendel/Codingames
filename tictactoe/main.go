package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"
)

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
			// Calculate the correct indices for the sub-board
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

		// Calculate the sub-board to play in
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
		// Output the chosen move in terms of the global board
		globalRow := subBoardRow*boardSize + bm.Row
		globalCol := subBoardCol*boardSize + bm.Col
		fmt.Fprintf(os.Stderr, "glob row %2d, col %2d\n", globalRow, globalCol)
		// Validate the selected move
		if !isValidMove(validMoves, Move{Row: globalRow, Col: globalCol}) {
			fmt.Fprintln(os.Stderr, "ERROR: MCTS selected an invalid move")
			bm = validMoves[0]
		}

		fmt.Println(globalRow, globalCol)
	}
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
