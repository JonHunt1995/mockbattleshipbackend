package ship

import (
	"errors"
	"strings"
)

var ErrInvalidShipPlacement = errors.New("The uploaded ships are not properly placed")

type ShipCoordinates struct {
	Carrier    []int `json:"Carrier"`
	Battleship []int `json:"Battleship"`
	Cruiser    []int `json:"Cruiser"`
	Submarine  []int `json:"Submarine"`
	Destroyer  []int `json:"Destroyer"`
}

type LivingShips struct {
	Carrier    bool
	Battleship bool
	Cruiser    bool
	Submarine  bool
	Destroyer  bool
}

func NewLivingShips() LivingShips {
	return LivingShips{
		Carrier:    true,
		Battleship: true,
		Cruiser:    true,
		Submarine:  true,
		Destroyer:  true,
	}
}

func (sc *ShipCoordinates) GetFlattenedCoords() []int {
	const lengthOfAllShips = 17
	flattened := make([]int, 0, lengthOfAllShips)

	flattened = append(flattened, sc.Carrier...)
	flattened = append(flattened, sc.Battleship...)
	flattened = append(flattened, sc.Cruiser...)
	flattened = append(flattened, sc.Submarine...)
	flattened = append(flattened, sc.Destroyer...)

	return flattened
}

func (sc *ShipCoordinates) AreValid() error {
	if validShipPlacement(sc.Carrier, "carrier") &&
		validShipPlacement(sc.Battleship, "battleship") &&
		validShipPlacement(sc.Cruiser, "cruiser") &&
		validShipPlacement(sc.Submarine, "submarine") &&
		validShipPlacement(sc.Destroyer, "destroyer") &&
		noOverlaps(sc.GetFlattenedCoords()) {
		return nil
	}

	return ErrInvalidShipPlacement
}

func validShipPlacement(s []int, name string) bool {
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

	if len(s) < 2 || lengthOfEachShip[name] != len(s) {
		return false
	}

	diff := s[1] - s[0]

	if diff != 1 && diff != 10 {
		return false
	}

	for i, idx := range s {
		if 0 > idx || idx >= 100 {
			return false
		}
		if i != 0 && idx-s[i-1] != diff {
			return false
		}
		if diff == 1 && idx/10 != s[0]/10 {
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
