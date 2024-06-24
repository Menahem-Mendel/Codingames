package main

type Score map[Medal]int

func NewScore(gold, silver, bronze int) Score {
	return Score{
		GOLD:   gold,
		SILVER: silver,
		BRONZE: bronze,
	}
}

func (s Score) Calc() int {
	return s[GOLD]*3 + s[SILVER]
}
