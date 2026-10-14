package game

import (
	"errors"
	"testing"

	"github.com/JonHunt1995/mockbattleshipbackend/player"
	"github.com/JonHunt1995/mockbattleshipbackend/ship"
)

func TestGame_ValidateTurn(t *testing.T) {
	p1 := &player.Player{Id: "p1-uuid"}
	p2 := &player.Player{Id: "p2-uuid"}
	g := NewGame([]*player.Player{p1, p2})

	tests := []struct {
		name     string
		setTurn  int
		playerID string
		wantErr  bool
	}{
		// P1 is valid on Odd turns
		{"Turn 1: P1's turn (Valid)", 1, "p1-uuid", false},
		{"Turn 1: P2 tries to move (Invalid)", 1, "p2-uuid", true},
		{"Turn 3: P1's turn again (Valid)", 3, "p1-uuid", false},

		// P2 is valid on Even turns
		{"Turn 2: P2's turn (Valid)", 2, "p2-uuid", false},
		{"Turn 2: P1 tries to move (Invalid)", 2, "p1-uuid", true},
		{"Turn 4: P2's turn again (Valid)", 4, "p2-uuid", false},

		{"Non-existent player moves (Invalid)", 1, "hacker-id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1.Guesses = nil
			p2.Guesses = nil
			for i := 0; i < tt.setTurn-1; i++ {
				if i%2 == 0 {
					p1.Guesses = append(p1.Guesses, i)
				} else {
					p2.Guesses = append(p2.Guesses, i)
				}
			}
			err := g.ValidateTurn(tt.playerID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTurn(%q) error presence:\n  got error: %v\n  want error: %v", tt.playerID, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if !errors.Is(err, ErrNotYourTurn) {
					t.Errorf("ValidateTurn(%q) got error = %v, want %v", tt.playerID, err, ErrNotYourTurn)
				}
			}
		})
	}
}

func TestGame_FullMatchToVictory(t *testing.T) {
	p1Coords := ship.ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}
	p2Coords := ship.ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}

	p1 := player.NewPlayer(p1Coords, "player-1")
	p2 := player.NewPlayer(p2Coords, "player-2")
	g := NewGame([]*player.Player{p1, p2})

	p2Targets := p2Coords.GetFlattenedCoords()
	if len(p2Targets) != 17 {
		t.Fatalf("expected 17 target coordinates for player 2, got %d", len(p2Targets))
	}

	// Play 17 rounds: Player 1 hits every one of Player 2's ships; Player 2 misses at 80+i
	for i, target := range p2Targets {
		if err := g.PlayTurn("player-1", target); err != nil {
			t.Fatalf("Turn %d: player-1 failed to fire at target %d: %v", g.GetTurnNumber(), target, err)
		}

		if i == len(p2Targets)-1 {
			break
		}

		midState, err := g.GetGameState("player-1")
		if err != nil {
			t.Fatalf("failed to get mid-game state: %v", err)
		}
		if midState.Victor != 0 {
			t.Fatalf("Turn %d: expected Victor == 0 mid-game, got %d", g.GetTurnNumber(), midState.Victor)
		}

		missCoord := 80 + i
		if err := g.PlayTurn("player-2", missCoord); err != nil {
			t.Fatalf("Turn %d: player-2 failed to fire at %d: %v", g.GetTurnNumber(), missCoord, err)
		}
	}

	p1State, err := g.GetGameState("player-1")
	if err != nil {
		t.Fatalf("failed to get player 1 game state: %v", err)
	}

	if p1State.Victor != 1 {
		t.Errorf("expected player 1 Victor == 1, got %d", p1State.Victor)
	}
	if !p1State.GameIsReady {
		t.Errorf("expected GameIsReady == true")
	}
	if len(p1State.PlayerHits) != 17 {
		t.Errorf("expected 17 player hits scored, got %d", len(p1State.PlayerHits))
	}
	if len(p1State.OpponentHits) != 0 {
		t.Errorf("expected 0 opponent hits scored on player 1, got %d", len(p1State.OpponentHits))
	}
	if len(p1State.OpponentMisses) != 16 {
		t.Errorf("expected 16 opponent misses scored on player 1, got %d", len(p1State.OpponentMisses))
	}

	p2Living := p1State.OpponentLivingShips
	checkAllShipsSunk(t, p2Living)

	p1Living := p1State.PlayerLivingShips
	checkAllShipsAlive(t, p1Living)

	p2State, err := g.GetGameState("player-2")
	if err != nil {
		t.Fatalf("failed to get player 2 game state: %v", err)
	}

	if p2State.Victor != -1 {
		t.Errorf("expected player 2 Victor == -1, got %d", p2State.Victor)
	}
}

func TestGame_VictoryPermutations_BothPlayers(t *testing.T) {
	shipTypes := []string{"Carrier", "Battleship", "Cruiser", "Submarine", "Destroyer"}
	players := []struct {
		winnerID string
		loserID  string
	}{
		{"player-1", "player-2"},
		{"player-2", "player-1"},
	}

	for _, p := range players {
		for _, lastShip := range shipTypes {
			testName := p.winnerID + "_wins_with_" + p.loserID + "_" + lastShip + "_last"
			t.Run(testName, func(t *testing.T) {
				p1Coords := ship.ShipCoordinates{
					Carrier:    []int{0, 1, 2, 3, 4},
					Battleship: []int{10, 20, 30, 40},
					Cruiser:    []int{50, 51, 52},
					Submarine:  []int{60, 61, 62},
					Destroyer:  []int{70, 71},
				}
				p2Coords := ship.ShipCoordinates{
					Carrier:    []int{0, 1, 2, 3, 4},
					Battleship: []int{10, 20, 30, 40},
					Cruiser:    []int{50, 51, 52},
					Submarine:  []int{60, 61, 62},
					Destroyer:  []int{70, 71},
				}

				p1 := player.NewPlayer(p1Coords, "player-1")
				p2 := player.NewPlayer(p2Coords, "player-2")
				g := NewGame([]*player.Player{p1, p2})

				shipsMap := map[string][]int{
					"Carrier":    []int{0, 1, 2, 3, 4},
					"Battleship": []int{10, 20, 30, 40},
					"Cruiser":    []int{50, 51, 52},
					"Submarine":  []int{60, 61, 62},
					"Destroyer":  []int{70, 71},
				}

				// Deterministic ordering: 4 ships in firstTargets, remaining lastShip in finalTargets
				order := []string{"Carrier", "Battleship", "Cruiser", "Submarine", "Destroyer"}
				var firstTargets []int
				var finalTargets []int

				for _, name := range order {
					coords := shipsMap[name]
					if name == lastShip {
						finalTargets = append(finalTargets, coords...)
					} else {
						firstTargets = append(firstTargets, coords...)
					}
				}

				allTargets := append(firstTargets, finalTargets...)
				if len(allTargets) != 17 {
					t.Fatalf("expected 17 targets, got %d", len(allTargets))
				}

				// Execute turns:
				// When player-1 wins: P1 fires hits on odd turns, P2 fires misses on even turns.
				// When player-2 wins: P1 fires misses on odd turns, P2 fires hits on even turns.
				for i, target := range allTargets {
					if p.winnerID == "player-1" {
						// P1 (winner) hits target
						if err := g.PlayTurn("player-1", target); err != nil {
							t.Fatalf("Turn %d: P1 failed to hit %d: %v", g.GetTurnNumber(), target, err)
						}

						// Mid-game check: exactly when all firstTargets (4 ships) have been sunk
						if i == len(firstTargets)-1 {
							verifyMidGameState(t, g, p.winnerID, p.loserID, lastShip)
						}

						// P2 (loser) fires miss (80+i), unless match just ended
						if i < len(allTargets)-1 {
							miss := 80 + i
							if err := g.PlayTurn("player-2", miss); err != nil {
								t.Fatalf("Turn %d: P2 failed to fire miss %d: %v", g.GetTurnNumber(), miss, err)
							}
						}
					} else {
						// P1 (loser) fires miss (80+i)
						miss := 80 + i
						if err := g.PlayTurn("player-1", miss); err != nil {
							t.Fatalf("Turn %d: P1 failed to fire miss %d: %v", g.GetTurnNumber(), miss, err)
						}

						// P2 (winner) hits target
						if err := g.PlayTurn("player-2", target); err != nil {
							t.Fatalf("Turn %d: P2 failed to hit %d: %v", g.GetTurnNumber(), target, err)
						}

						// Mid-game check: exactly when all firstTargets (4 ships) have been sunk
						if i == len(firstTargets)-1 {
							verifyMidGameState(t, g, p.winnerID, p.loserID, lastShip)
						}
					}
				}

				// Final check: All 17 targets hit, verify victory from both viewpoints
				verifyFinalGameState(t, g, p.winnerID, p.loserID)
			})
		}
	}
}

func verifyMidGameState(t *testing.T, g *Game, winnerID, loserID, lastShip string) {
	t.Helper()

	// 1. Winner's perspective mid-game
	winnerState, err := g.GetGameState(winnerID)
	if err != nil {
		t.Fatalf("failed to get winner mid-game state: %v", err)
	}
	if winnerState.Victor != 0 {
		t.Errorf("mid-game: expected winner Victor == 0, got %d", winnerState.Victor)
	}
	checkAllShipsAlive(t, winnerState.PlayerLivingShips)
	checkOnlyOneShipAlive(t, winnerState.OpponentLivingShips, lastShip)

	// 2. Loser's perspective mid-game
	loserState, err := g.GetGameState(loserID)
	if err != nil {
		t.Fatalf("failed to get loser mid-game state: %v", err)
	}
	if loserState.Victor != 0 {
		t.Errorf("mid-game: expected loser Victor == 0, got %d", loserState.Victor)
	}
	checkAllShipsAlive(t, loserState.OpponentLivingShips)
	checkOnlyOneShipAlive(t, loserState.PlayerLivingShips, lastShip)
}

func verifyFinalGameState(t *testing.T, g *Game, winnerID, loserID string) {
	t.Helper()

	// 1. Winner's perspective final
	winnerState, err := g.GetGameState(winnerID)
	if err != nil {
		t.Fatalf("failed to get winner final game state: %v", err)
	}
	if winnerState.Victor != 1 {
		t.Errorf("final: expected winner Victor == 1, got %d", winnerState.Victor)
	}
	checkAllShipsAlive(t, winnerState.PlayerLivingShips)
	checkAllShipsSunk(t, winnerState.OpponentLivingShips)

	// 2. Loser's perspective final
	loserState, err := g.GetGameState(loserID)
	if err != nil {
		t.Fatalf("failed to get loser final game state: %v", err)
	}
	if loserState.Victor != -1 {
		t.Errorf("final: expected loser Victor == -1, got %d", loserState.Victor)
	}
	checkAllShipsAlive(t, loserState.OpponentLivingShips)
	checkAllShipsSunk(t, loserState.PlayerLivingShips)
}

func checkOnlyOneShipAlive(t *testing.T, living ship.LivingShips, expectedShip string) {
	t.Helper()
	ships := map[string]bool{
		"Carrier":    living.Carrier,
		"Battleship": living.Battleship,
		"Cruiser":    living.Cruiser,
		"Submarine":  living.Submarine,
		"Destroyer":  living.Destroyer,
	}
	for name, alive := range ships {
		if name == expectedShip {
			if !alive {
				t.Errorf("mid-game: expected only %s to be alive, but was sunk", name)
			}
		} else {
			if alive {
				t.Errorf("mid-game: expected %s to be sunk, but was alive", name)
			}
		}
	}
}

func checkAllShipsAlive(t *testing.T, living ship.LivingShips) {
	t.Helper()
	if !living.Carrier || !living.Battleship || !living.Cruiser || !living.Submarine || !living.Destroyer {
		t.Errorf("expected all ships to be alive, got: %+v", living)
	}
}

func checkAllShipsSunk(t *testing.T, living ship.LivingShips) {
	t.Helper()
	if living.Carrier || living.Battleship || living.Cruiser || living.Submarine || living.Destroyer {
		t.Errorf("expected all ships to be sunk, got: %+v", living)
	}
}
