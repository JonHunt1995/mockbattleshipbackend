package main

import "slices"

type Player struct {
	Ships   ShipCoordinates
	Id      string
	Guesses []int
}

type LivingShips struct {
	Carrier    bool
	Battleship bool
	Cruiser    bool
	Submarine  bool
	Destroyer  bool
}

func NewLivingShips() *LivingShips {
	return &LivingShips{
	Carrier: true,
	Battleship: true,
	Cruiser: true,
	Submarine: true,
	Destroyer: true
	}


func NewPlayer(sc ShipCoordinates, id string) *Player {
	var guesses []int

	return &Player{
		Ships:   sc,
		Id:      id,
		Guesses: guesses,
	}
}
func (p *Player) validateGuess(guess int) error {
	if guess < 0 || guess >= 100 {
		return &badRequest{
			status: 400, // Bad Request
			msg:    "This guess is out of bounds from the grid",
		}
	}
	if slices.Contains(p.Guesses, guess) {
		return &badRequest{
			status: 400, // Bad Request
			msg:    "You already played this",
		}
	}

	return nil
}

func (p *Player) AddGuess(guess int) error {
	if err := p.validateGuess(guess); err != nil {
		return err
	}

	p.Guesses = append(p.Guesses, guess)

	return nil
}

func (p *Player) getHitsAndMisses(other *Player) ([]int, []int) {
	hitMap := make(map[int]bool)
	var hits, misses []int

	ships := other.Ships.getFlattenedCoords()

	for _, ship := range ships {
		hitMap[ship] = true
	}

	for _, guess := range p.Guesses {
		if hitMap[guess] {
			hits = append(hits, guess)
			continue
		}

		misses = append(misses, guess)
	}

	return hits, misses
}

func (p *Player) getLivingShips(other *Player) LivingShips {
	guesses := make(map[int]bool)
	for _, guess := range other.Guesses {
		guesses[guess] = true
	}

	return LivingShips{
		Carrier:    isShipAlive(p.Ships.Carrier, guesses),
		Battleship: isShipAlive(p.Ships.Battleship, guesses),
		Cruiser:    isShipAlive(p.Ships.Cruiser, guesses),
		Submarine:  isShipAlive(p.Ships.Submarine, guesses),
		Destroyer:  isShipAlive(p.Ships.Destroyer, guesses),
	}
}

func isShipAlive(coords []int, guesses map[int]bool) bool {
	for _, coord := range coords {
		if !guesses[coord] {
			return true
		}
	}

	return false
}
