package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	var playerIdx int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &playerIdx)

	var nbGames int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &nbGames)

	engine := NewEngine(
		playerIdx,
		NewHurdling(NewHurdler(), NewHurdler(), NewHurdler()),
		NewArchery(NewArcher(), NewArcher(), NewArcher()),
		NewSkating(NewSkater(), NewSkater(), NewSkater()),
		NewDiving(NewDiver(), NewDiver(), NewDiver()),
	)

	engine.ListenAndServe(scanner)
}
