# Arcade Olympics Bot

## Goal
The objective of the game is to end with a higher score than your opponents. Three players compete against each other in the arcade Olympics, each controlling a character in four mini-games simultaneously. Earn the maximum number of medals in all four games to acquire the highest score.

## Rules
Each player is connected to four different arcade machines, each running a different mini-game. The code can read the 8 registers used internally by the machines: `GPU`, containing a string, and `reg0` to `reg6` containing integers. The meaning of these values varies by game.

The game is played in turns. On each turn, all three players perform one of four possible actions: `UP`, `DOWN`, `LEFT`, or `RIGHT`. When an action is performed, their agents in each mini-game perform the same action simultaneously.

### Earning Medals
The four mini-games play on loop throughout the game. In each run of a mini-game, you may acquire a gold, silver, or bronze medal. Between runs, there is a reset turn where the mini-game is inactive.

At the end of the game, each player's score for each mini-game is calculated based on the number of medals earned, using the following formula:
```
mini_game_score = nb_silver_medals + nb_gold_medals * 3
```
The scores for all four mini-games are multiplied together to determine the final score.

During a reset turn, the `GPU` register will show `GAME_OVER`.

If there are ties in a mini-game, tied players will win the same highest medal. For instance, if two players tie for first place, they will both win gold and the third player will receive bronze.

## Mini-Games

### Mini-game 1: Hurdle Race
This is a race between the three agents on a randomly generated track of 30 spaces. Each space may contain a hurdle, which agents must jump over or be stunned for the next 3 turns.

**Actions:**
- `UP`: Jump over one space, ignoring any hurdle on the next space, and move by 2 spaces total.
- `LEFT`: Move forward by 1 space.
- `DOWN`: Move forward by 2 spaces.
- `RIGHT`: Move forward by 3 spaces.

**Registers:**
- `GPU`: ASCII representation of the racetrack (`.` for empty space, `#` for hurdle).
- `reg0`, `reg1`, `reg2`: Positions of players 1, 2, and 3, respectively.
- `reg3`, `reg4`, `reg5`: Stun timers for players 1, 2, and 3, respectively.
- `reg6`: Unused.

### Mini-game 2: Archery
Players control a cursor with x and y coordinates. Each turn, players pick a direction and move their cursor by the current wind strength in that direction. After 12-15 turns, players win medals based on how close they are to (0,0).

**Registers:**
- `GPU`: Series of integers indicating wind strength for upcoming turns. The integer at index 0 is the current wind strength.
- `reg0`, `reg1`: x and y coordinates for player 1.
- `reg2`, `reg3`: x and y coordinates for player 2.
- `reg4`, `reg5`: x and y coordinates for player 3.
- `reg6`: Unused.

### Mini-game 3: Roller Speed Skating
Players race on a cyclical track of 10 spaces, with a risk attribute ranging from 0 to 5.

**Actions:**
- The order of the actions is provided in the `GPU` each turn, affecting the risk and movement.
- Actions at higher indices move the player more spaces but increase risk.

**Registers:**
- `GPU`: This turn's risk order (e.g., `ULDR`).
- `reg0`, `reg1`, `reg2`: Spaces traveled by players 1, 2, and 3, respectively.
- `reg3`, `reg4`, `reg5`: Risk or stun timers for players 1, 2, and 3, respectively.
- `reg6`: Turns left.

### Mini-game 4: Diving
Players must match the sequence of directions given at the start of each run (diving goal). Each matching action increments the combo multiplier, earning points equal to its value.

**Registers:**
- `GPU`: This run's diving goal (e.g., `UUUDDLLLULDRLL`).
- `reg0`, `reg1`, `reg2`: Points for players 1, 2, and 3, respectively.
- `reg3`, `reg4`, `reg5`: Combos for players 1, 2, and 3, respectively.
- `reg6`: Unused.

## Victory Condition
You win if you have a higher final score after 100 turns.

## Defeat Condition
You lose if your program does not provide a command in the allotted time or provides an unrecognized command.

## Game Protocol

### Initialization Input
- First line: `playerIdx` - an integer indicating which agent you control in the mini-games.
- Next line: The number of simultaneously running mini-games. For this league, it's 4.

### Input for One Game Turn
- Next 3 lines: One line per player, ordered by `playerIdx`. A string `scoreInfo` containing a breakdown of each player's final score. It contains 13 integers:
  - The first integer represents the player's current final score points.
  - The next three integers represent the number of gold, silver, and bronze medals for each mini-game.

- Next `nbGames` lines: One line for each mini-game, containing the eight space-separated registers:
  ```
  gpu a string
  reg0 an integer
  reg1 an integer
  reg2 an integer
  reg3 an integer
  reg4 an integer
  reg5 an integer
  reg6 an integer
  ```
  Their values depend on the game. Unused registers will always be -1.

### Output
One of the following strings:
- `UP`
- `RIGHT`
- `DOWN`
- `LEFT`

The effect will depend on the game.

### Constraints
- `0 ≤ playerIdx ≤ 2`
- `1 ≤ nbGames ≤ 4` (across all leagues)

- Response time per turn ≤ 50ms
- Response time for the first turn ≤ 1000ms

## Debugging Tips
- Press the gear icon on the viewer to access extra display options.
- Use the keyboard to control the action: space to play/pause, arrows to step one frame at a time.

---

By following this README, you should be able to understand the game's rules, the structure of the input and output, and how to implement a bot to compete in the arcade Olympics effectively.