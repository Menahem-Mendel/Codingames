package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
)

type Engine struct {
	teamTotal [3]int
	playerIdx int
	races     map[string]Game
}

func NewEngine(playerIdx int, games ...Game) Engine {
	races := make(map[string]Game, len(games))

	for _, game := range games {
		switch game.(type) {
		case *Hurdling:
			races[HURDLING] = game
		case *Diving:
			races[DIVING] = game
		case *Skating:
			races[SKATING] = game
		case *Archery:
			races[ARCHERY] = game
		default:
			fmt.Fprintln(os.Stderr, "UNKNOWN TYPE OF GAME:", game)
		}
	}

	return Engine{
		races:     races,
		playerIdx: playerIdx,
	}
}

func (e Engine) total(idx int) int { return e.teamTotal[idx] }

func (e Engine) ListenAndServe(scanner *bufio.Scanner) {
	nbPlayers := 3

	for {
		for i := 0; i < nbPlayers; i++ {
			scanner.Scan()
			scoreInfo := strings.Fields(scanner.Text())

			hurdlingScore := Score{
				GOLD:   toInt(scoreInfo[1]),
				SILVER: toInt(scoreInfo[2]),
				BRONZE: toInt(scoreInfo[3]),
			}
			archeryScore := Score{
				GOLD:   toInt(scoreInfo[4]),
				SILVER: toInt(scoreInfo[5]),
				BRONZE: toInt(scoreInfo[6]),
			}
			skatingScore := Score{
				GOLD:   toInt(scoreInfo[7]),
				SILVER: toInt(scoreInfo[8]),
				BRONZE: toInt(scoreInfo[9]),
			}
			divingScore := Score{
				GOLD:   toInt(scoreInfo[10]),
				SILVER: toInt(scoreInfo[11]),
				BRONZE: toInt(scoreInfo[12]),
			}

			e.teamTotal[i] = toInt(scoreInfo[0])
			UpdatePlayer(e.races[HURDLING].Player(i), hurdlingScore)
			UpdatePlayer(e.races[ARCHERY].Player(i), archeryScore)
			UpdatePlayer(e.races[SKATING].Player(i), skatingScore)
			UpdatePlayer(e.races[DIVING].Player(i), divingScore)
		}

		gpu, regs := ParseState(scanner)
		UpdateGame(e.races[HURDLING], gpu, regs)
		gpu, regs = ParseState(scanner)
		UpdateGame(e.races[ARCHERY], gpu, regs)
		gpu, regs = ParseState(scanner)
		UpdateGame(e.races[SKATING], gpu, regs)
		gpu, regs = ParseState(scanner)
		UpdateGame(e.races[DIVING], gpu, regs)

		action := e.Exec()

		fmt.Println(action)
	}
}
func geometricMean(totalScore int, numGames int) float64 {
	if totalScore <= 0 {
		return 0
	}
	return math.Pow(float64(totalScore), 1.0/float64(numGames))
}

func (e Engine) Exec() Command {
	bestAction := LEFT
	maxBias := -1 << 31

	// Calculate the geometric mean using the total score
	geomMean := geometricMean(e.total(e.playerIdx), len(e.races))

	for _, cmd := range []Command{UP, DOWN, LEFT, RIGHT} {
		totalBias := 0

		for key, game := range e.races {
			bias := game.Eval(cmd, e.playerIdx)
			playerScore := game.Player(e.playerIdx).Score()
			place := game.Place(game.Player(e.playerIdx))

			// // Calculate score importance based on geometric mean
			if float64(playerScore) <= geomMean {
				// Prioritize games with scores below the geometric mean
				bias = int(float64(bias) * 5)
			} else {
				// Deprioritize games with scores above the geometric mean
				bias = int(float64(bias) * 0.01)
			}

			// Adjust bias based on potential to win or improve in the game
			opponentScores := []int{}
			for i := 0; i < 3; i++ {
				if i != e.playerIdx {
					opponentScores = append(opponentScores, game.Player(i).Score())
				}
			}

			// Consider the highest opponent score in the game
			maxOpponentScore := max(opponentScores)

			// If player's score is significantly lower than the highest opponent score, deprioritize the game
			if playerScore < maxOpponentScore-10 {
				bias = int(float64(bias) * 0.5)
			}

			// Prioritize games where the player is in the highest place
			if place == 1 {
				bias = int(float64(bias) * 3)
			} else if place == 2 {
				bias = int(float64(bias) * 1.5)
			} else if place == 3 {
				bias *= 5
			}

			// if bias > 0 {
			fmt.Fprintf(os.Stderr, "GAME: %8s, ACTION: %5s, BIAS: %5d, PLAYER SCORE: %3d, PLACE: %d, GEOM MEAN: %.2f\n", key, cmd, bias, playerScore, place, geomMean)
			// }
			totalBias += bias
		}

		if totalBias > maxBias {
			maxBias = totalBias
			bestAction = cmd
		}
	}

	return bestAction
}

// Helper function to find the maximum value in a slice
func max(slice []int) int {
	maxValue := slice[0]
	for _, value := range slice {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}
