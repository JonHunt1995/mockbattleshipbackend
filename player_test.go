package main

import (
	"reflect"
	"testing"
)

func TestPlayer_Guesses(t *testing.T) {
	// Setup a player with some initial state
	p := &Player{
		id:      "test-player",
		guesses: []int{10, 20},
	}

	// Table for validateGuess and AddGuess
	tests := []struct {
		name      string
		guess     int
		wantError bool
		errorMsg  string
	}{
		{"Valid Guess", 30, false, ""},
		{"Valid Boundary 0", 0, false, ""},
		{"Valid Boundary 99", 99, false, ""},
		{"Error: Negative", -1, true, "This guess is out of bounds from the grid"},
		{"Error: Too High", 100, true, "This guess is out of bounds from the grid"},
		{"Error: Already Played", 10, true, "You already played this"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Test validateGuess
			err := p.validateGuess(tt.guess)
			if (err != nil) != tt.wantError {
				t.Errorf("validateGuess(%d) error = %v, wantError %v", tt.guess, err, tt.wantError)
			}
			if tt.wantError && err != nil && err.Error() != tt.errorMsg {
				t.Errorf("validateGuess() error message = %q, want %q", err.Error(), tt.errorMsg)
			}

			// 2. Test AddGuess
			// We only call AddGuess if we want to test the success/failure flow
			initialCount := len(p.guesses)
			err = p.AddGuess(tt.guess)

			if !tt.wantError {
				if err != nil {
					t.Errorf("AddGuess(%d) failed unexpectedly: %v", tt.guess, err)
				}
				if len(p.guesses) != initialCount+1 {
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
		otherShips ShipCoordinates
		wantHits   []int
		wantMisses []int
	}{
		{
			name:      "All Misses",
			myGuesses: []int{1, 2, 3},
			otherShips: ShipCoordinates{
				Destroyer: []int{50, 60},
			},
			wantHits:   nil,
			wantMisses: []int{1, 2, 3},
		},
		{
			name:      "All Hits",
			myGuesses: []int{50, 60},
			otherShips: ShipCoordinates{
				Destroyer: []int{50, 60},
			},
			wantHits:   []int{50, 60},
			wantMisses: nil,
		},
		{
			name:      "Mixed Hits and Misses",
			myGuesses: []int{10, 11, 12},
			otherShips: ShipCoordinates{
				Cruiser: []int{11, 21, 31},
			},
			wantHits:   []int{11},
			wantMisses: []int{10, 12},
		},
		{
			name:       "Empty Guesses",
			myGuesses:  []int{},
			otherShips: ShipCoordinates{Destroyer: []int{1, 2}},
			wantHits:   nil,
			wantMisses: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{guesses: tt.myGuesses}
			other := &Player{ships: tt.otherShips}

			gotHits, gotMisses := p.getHitsAndMisses(other)

			if !reflect.DeepEqual(gotHits, tt.wantHits) {
				t.Errorf("getHitsAndMisses() gotHits = %v, want %v", gotHits, tt.wantHits)
			}
			if !reflect.DeepEqual(gotMisses, tt.wantMisses) {
				t.Errorf("getHitsAndMisses() gotMisses = %v, want %v", gotMisses, tt.wantMisses)
			}
		})
	}
}
