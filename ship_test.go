package main

import (
	"reflect"
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

func TestNoOverlaps(t *testing.T) {
	tests := []struct {
		name   string
		coords []int
		want   bool
	}{
		{"Empty slice", []int{}, true},
		{"No duplicates", []int{1, 2, 3, 10, 20}, true},
		{"Simple duplicate", []int{1, 2, 1}, false},
		{"Large duplicate", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, false},
		{"Negative duplicates", []int{-1, -1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := noOverlaps(tt.coords); got != tt.want {
				t.Errorf("noOverlaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFlattenedCoords(t *testing.T) {
	tests := []struct {
		name string
		sc   ShipCoordinates
		want []int
	}{
		{
			name: "Happy Path - All Ships",
			sc: ShipCoordinates{
				Carrier:    []int{1},
				Battleship: []int{2},
				Cruiser:    []int{3},
				Submarine:  []int{4},
				Destroyer:  []int{5},
			},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "Partial Board - Only Destroyer",
			sc: ShipCoordinates{
				Destroyer: []int{98, 99},
			},
			want: []int{98, 99},
		},
		{
			name: "Empty Struct - All Nil",
			sc:   ShipCoordinates{},
			want: []int{},
		},
		{
			name: "Irregular Lengths (Too many/too few)",
			sc: ShipCoordinates{
				Carrier:   []int{1, 2}, // Should be 5, but flattener shouldn't care
				Destroyer: []int{10, 11, 12, 13, 14, 15},
			},
			want: []int{1, 2, 10, 11, 12, 13, 14, 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sc.getFlattenedCoords()

			// If both are empty, DeepEqual might fail if one is nil and other is []int{}
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestShipCoordinates_AreValid(t *testing.T) {
	tests := []struct {
		name string
		sc   ShipCoordinates
		want bool
	}{
		{
			name: "Perfect Board (All Valid)",
			sc: ShipCoordinates{
				Carrier:    []int{0, 1, 2, 3, 4},
				Battleship: []int{10, 20, 30, 40},
				Cruiser:    []int{97, 98, 99},
				Submarine:  []int{50, 51, 52},
				Destroyer:  []int{70, 80},
			},
			want: true,
		},
		{
			name: "Overlap Failure (Carrier and Battleship touch at 10)",
			sc: ShipCoordinates{
				Carrier:    []int{10, 11, 12, 13, 14},
				Battleship: []int{0, 10, 20, 30}, // Overlap at 10
				Cruiser:    []int{97, 98, 99},
				Submarine:  []int{50, 51, 52},
				Destroyer:  []int{70, 80},
			},
			want: false,
		},
		{
			name: "Individual Ship Failure (Battleship too short)",
			sc: ShipCoordinates{
				Carrier:    []int{0, 1, 2, 3, 4},
				Battleship: []int{10, 11, 12}, // Should be 4, is 3
				Cruiser:    []int{20, 21, 22},
				Submarine:  []int{30, 31, 32},
				Destroyer:  []int{40, 41},
			},
			want: false,
		},
		{
			name: "Horizontal Wrap Failure",
			sc: ShipCoordinates{
				Carrier:    []int{0, 1, 2, 3, 4},
				Battleship: []int{10, 11, 12, 13},
				Cruiser:    []int{20, 21, 22},
				Submarine:  []int{30, 31, 32},
				Destroyer:  []int{9, 10}, // Invalid wrap from row 0 to 1
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sc.areValid(); got != tt.want {
				t.Errorf("areValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
