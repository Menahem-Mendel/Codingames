package main

// Strategy interface represents a strategy for determining actions.
// It allows different strategies to be implemented and applied to games.
type Strategy interface {
	// Apply the strategy to the game and return the chosen action.
	Apply(game Game) string
}

// StrategyFunc is a function type that implements the Strategy interface.
// It allows a function to be used as a strategy by implementing the Apply method.
type StrategyFunc func(g Game) string

// Apply executes the strategy function on the game and returns the chosen action.
func (sf StrategyFunc) Apply(game Game) string {
	return sf(game)
}
