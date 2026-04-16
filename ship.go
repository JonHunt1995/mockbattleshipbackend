package main

import (
	"fmt"
	"strings"
)

type ShipCoordinates struct {
	Carrier    []int `json:"Carrier"`
	Battleship []int `json:"Battleship"`
	Cruiser    []int `json:"Cruiser"`
	Submarine  []int `json:"Submarine"`
	Destroyer  []int `json:"Destroyer"`
}

func (sc *ShipCoordinates) getFlattenedCoords() []int {
	const lengthOfAllShips = 17
	flattened := make([]int, 0, lengthOfAllShips)

	flattened = append(flattened, sc.Carrier...)
	flattened = append(flattened, sc.Battleship...)
	flattened = append(flattened, sc.Cruiser...)
	flattened = append(flattened, sc.Submarine...)
	flattened = append(flattened, sc.Destroyer...)

	return flattened
}

func (sc *ShipCoordinates) areValid() error {
	if validShipPlacement(sc.Carrier, "carrier") &&
		validShipPlacement(sc.Battleship, "battleship") &&
		validShipPlacement(sc.Cruiser, "cruiser") &&
		validShipPlacement(sc.Submarine, "submarine") &&
		validShipPlacement(sc.Destroyer, "destroyer") &&
		noOverlaps(sc.getFlattenedCoords()) {
		return nil
	}

	return fmt.Errorf("The uploaded ships are not properly placed")
}

func validShipPlacement(ship []int, name string) bool {
	// Check if length of ship is correct.
	// Check if horizontal or vertical (diff should be
	// 1 or 10 respectively).
	// Check if any index is out of bounds (not between
	// 0 and 99).
	// Check if the gap between each index is
	// consistent throughout slice which means the ship
	// is contiguous and ensure if the ship is horizontal,
	// that the ship fits in the same row.
	lengthOfEachShip := map[string]int{
		"carrier":    5,
		"battleship": 4,
		"cruiser":    3,
		"submarine":  3,
		"destroyer":  2,
	}

	name = strings.ToLower(name)

	if len(ship) < 2 || lengthOfEachShip[name] != len(ship) {
		return false
	}

	diff := ship[1] - ship[0]

	if diff != 1 && diff != 10 {
		return false
	}

	for i, idx := range ship {
		if 0 > idx || idx >= 100 {
			return false
		}
		if i != 0 && idx-ship[i-1] != diff {
			return false
		}
		if diff == 1 && idx/10 != ship[0]/10 {
			return false
		}
	}

	return true
}

func noOverlaps(coords []int) bool {
	distinctCoordsMap := make(map[int]bool)

	for _, coord := range coords {
		if distinctCoordsMap[coord] {
			return false
		}
		distinctCoordsMap[coord] = true
	}

	return true
}
