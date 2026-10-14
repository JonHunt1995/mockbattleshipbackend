package player

import (
	"errors"
	"slices"

	"github.com/JonHunt1995/mockbattleshipbackend/ship"
)

var (
	ErrOutOfBounds   = errors.New("This guess is out of bounds from the grid")
	ErrAlreadyPlayed = errors.New("You already played this")
)

type Player struct {
	Ships   ship.ShipCoordinates
	Id      string
	Guesses []int
}

func NewPlayer(sc ship.ShipCoordinates, id string) *Player {
	var guesses []int

	return &Player{
		Ships:   sc,
		Id:      id,
		Guesses: guesses,
	}
}

func (p *Player) ValidateGuess(guess int) error {
	if guess < 0 || guess >= 100 {
		return ErrOutOfBounds
	}
	if slices.Contains(p.Guesses, guess) {
		return ErrAlreadyPlayed
	}

	return nil
}

func (p *Player) AddGuess(guess int) error {
	if err := p.ValidateGuess(guess); err != nil {
		return err
	}

	p.Guesses = append(p.Guesses, guess)

	return nil
}

func (p *Player) GetHitsAndMisses(other *Player) ([]int, []int) {
	hitMap := make(map[int]bool)
	var hits, misses []int

	ships := other.Ships.GetFlattenedCoords()

	for _, s := range ships {
		hitMap[s] = true
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

func (p *Player) GetLivingShips(other *Player) ship.LivingShips {
	guesses := make(map[int]bool)
	for _, guess := range other.Guesses {
		guesses[guess] = true
	}

	return ship.LivingShips{
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
