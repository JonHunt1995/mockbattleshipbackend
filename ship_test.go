package main

import (
	"testing"
)

func TestValidShipPlacement(t *testing.T) {
	tests := []struct {
		name     string
		ship     []int
		shipName string
		want     bool
	}{
		// --- VALID CASES ---
		{"Valid Carrier Horizontal", []int{0, 1, 2, 3, 4}, "carrier", true},
		{"Valid Battleship Vertical", []int{10, 20, 30, 40}, "battleship", true},
		{"Valid Destroyer Top Right", []int{8, 9}, "destroyer", true},
		{"Valid Destroyer Bottom Right", []int{98, 99}, "destroyer", true},
		{"Valid Submarine Mixed Case", []int{44, 45, 46}, "SubMaRiNe", true},

		// --- LENGTH & NAME FAILURES ---
		{"Invalid Name", []int{1, 2}, "tugboat", false},
		{"Wrong Length (Too Short)", []int{10, 20}, "cruiser", false},
		{"Wrong Length (Too Long)", []int{10, 11, 12}, "destroyer", false},
		{"Slice Too Short (Len 1)", []int{10}, "destroyer", false},
		{"Empty Slice", []int{}, "destroyer", false},

		// --- CONTIGUITY & ORIENTATION FAILURES ---
		{"Non-contiguous Horizontal", []int{1, 2, 4}, "cruiser", false},
		{"Non-contiguous Vertical", []int{10, 20, 40}, "cruiser", false},
		{"Diagonal (Diff 11)", []int{10, 21}, "destroyer", false},
		{"Diagonal (Diff 9)", []int{10, 19}, "destroyer", false},
		{"Reverse Order", []int{20, 10}, "destroyer", false},

		// --- BOUNDS FAILURES ---
		{"Negative Index", []int{-1, 0}, "destroyer", false},
		{"Index Too High", []int{99, 100}, "destroyer", false},

		// --- ROW WRAPPING FAILURES ---
		{"Horizontal Wrap Row 0 to 1", []int{9, 10}, "destroyer", false},
		{"Horizontal Wrap Row 4 to 5", []int{48, 49, 50}, "cruiser", false},
		{"Vertical 'Wrap' (Not possible with diff 10, but good to check)", []int{90, 100}, "destroyer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validShipPlacement(tt.ship, tt.shipName)
			if got != tt.want {
				t.Errorf("validShipPlacement(%v, %q) = %v; want %v", tt.ship, tt.shipName, got, tt.want)
			}
		})
	}
}
