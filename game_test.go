package main

import (
	"errors"
	"testing"
)

func TestGame_ValidateTurn(t *testing.T) {
	p1 := &Player{Id: "p1-uuid"}
	p2 := &Player{Id: "p2-uuid"}
	game := &Game{
		Players: []*Player{p1, p2},
		// Turn starts at 1
	}

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
			game.Turn = tt.setTurn
			err := game.validateTurn(tt.playerID)

			// 1. Check if the error presence matches expectations
			if (err != nil) != tt.wantErr {
				// Changed %d to %q since playerID is a string
				t.Errorf("validateTurn(%q) error presence:\n  got error: %v\n  want error: %v", tt.playerID, err, tt.wantErr)
				return
			}

			// 2. If an error was expected, verify the details
			if tt.wantErr {
				var br *badRequest
				if errors.As(err, &br) {
					if br.status != 400 {
						t.Errorf("validateTurn(%q) status:\n  got:  %d\n  want: %d", tt.playerID, br.status, 400)
					}
					if br.msg != "It is not your turn" {
						t.Errorf("validateTurn(%q) message:\n  got:  %q\n  want: %q", tt.playerID, br.msg, "It is not your turn")
					}
				} else {
					t.Errorf("validateTurn(%q) error type:\n  got:  %T\n  want: *badRequest", tt.playerID, err)
				}
			}
		})
	}
}
