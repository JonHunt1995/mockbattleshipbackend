package player

import (
	"errors"
	"slices"
	"testing"

	"github.com/JonHunt1995/mockbattleshipbackend/ship"
)

func TestPlayer_Guesses(t *testing.T) {
	// Setup a player with some initial state
	p := &Player{
		Id:      "test-player",
		Guesses: []int{10, 20},
	}

	tests := []struct {
		name      string
		guess     int
		wantError bool
		targetErr error
	}{
		{"Valid Guess", 30, false, nil},
		{"Valid Boundary 0", 0, false, nil},
		{"Valid Boundary 99", 99, false, nil},
		{"Error: Negative", -1, true, ErrOutOfBounds},
		{"Error: Too High", 100, true, ErrOutOfBounds},
		{"Error: Already Played", 10, true, ErrAlreadyPlayed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Test ValidateGuess
			err := p.ValidateGuess(tt.guess)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateGuess(%d) error = %v, wantError %v", tt.guess, err, tt.wantError)
			}
			if tt.wantError && !errors.Is(err, tt.targetErr) {
				t.Errorf("ValidateGuess(%d) error = %v, want %v", tt.guess, err, tt.targetErr)
			}

			// 2. Test AddGuess
			initialCount := len(p.Guesses)
			err = p.AddGuess(tt.guess)

			if !tt.wantError {
				if err != nil {
					t.Errorf("AddGuess(%d) failed unexpectedly: %v", tt.guess, err)
				}
				if len(p.Guesses) != initialCount+1 {
					t.Errorf("AddGuess(%d) did not increase guesses count", tt.guess)
				}
			} else if err == nil {
				t.Errorf("AddGuess(%d) should have failed but returned nil", tt.guess)
			}
		})
	}
}

func TestPlayer_GetHitsAndMisses(t *testing.T) {
	tests := []struct {
		name       string
		myGuesses  []int
		otherShips ship.ShipCoordinates
		wantHits   []int
		wantMisses []int
	}{
		{
			name:      "All Misses",
			myGuesses: []int{1, 2, 3},
			otherShips: ship.ShipCoordinates{
				Destroyer: []int{50, 60},
			},
			wantHits:   nil,
			wantMisses: []int{1, 2, 3},
		},
		{
			name:      "All Hits",
			myGuesses: []int{50, 60},
			otherShips: ship.ShipCoordinates{
				Destroyer: []int{50, 60},
			},
			wantHits:   []int{50, 60},
			wantMisses: nil,
		},
		{
			name:      "Mixed Hits and Misses",
			myGuesses: []int{10, 11, 12},
			otherShips: ship.ShipCoordinates{
				Cruiser: []int{11, 21, 31},
			},
			wantHits:   []int{11},
			wantMisses: []int{10, 12},
		},
		{
			name:       "Empty Guesses",
			myGuesses:  []int{},
			otherShips: ship.ShipCoordinates{Destroyer: []int{1, 2}},
			wantHits:   nil,
			wantMisses: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{Guesses: tt.myGuesses}
			other := &Player{Ships: tt.otherShips}

			gotHits, gotMisses := p.GetHitsAndMisses(other)

			if !slices.Equal(gotHits, tt.wantHits) {
				t.Errorf("GetHitsAndMisses() gotHits = %v, want %v", gotHits, tt.wantHits)
			}
			if !slices.Equal(gotMisses, tt.wantMisses) {
				t.Errorf("GetHitsAndMisses() gotMisses = %v, want %v", gotMisses, tt.wantMisses)
			}
		})
	}
}
