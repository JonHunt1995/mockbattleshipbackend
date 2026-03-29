package main

import (
	"errors"
	"testing"
)

func TestGame_ValidateTurn(t *testing.T) {
	// Setup a standard game state
	p1 := Player{id: "p1-uuid"}
	p2 := Player{id: "p2-uuid"}
	game := &Game{
		p1:   p1,
		p2:   p2,
		turn: 0,
	}

	tests := []struct {
		name     string
		setTurn  int
		playerID string
		wantErr  bool
	}{
		{"Turn 0: P1's turn (Valid)", 0, "p1-uuid", false},
		{"Turn 0: P2 tries to move (Invalid)", 0, "p2-uuid", true},
		{"Turn 1: P2's turn (Valid)", 1, "p2-uuid", false},
		{"Turn 1: P1 tries to move (Invalid)", 1, "p1-uuid", true},
		{"Turn 2: P1's turn again (Valid)", 2, "p1-uuid", false},
		{"Non-existent player moves (Invalid)", 0, "hacker-id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the game state
			game.turn = tt.setTurn

			// Execute the validation
			err := game.validateTurn(tt.playerID)

			// 1. Check if error presence matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTurn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 2. If an error occurred, verify it's our custom malformedRequest type
			if tt.wantErr {
				var mr *malformedRequest
				// errors.As checks type and unwrap at the same time
				if errors.As(err, &mr) {
					if mr.status != 400 {
						t.Errorf("Expected status 400, got %d", mr.status)
					}
					if mr.msg != "It is not your turn" {
						t.Errorf("Expected msg 'It is not your turn', got %q", mr.msg)
					}
				} else {
					t.Errorf("Error returned was not of type *malformedRequest")
				}
			}
		})
	}
}
