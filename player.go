package main

import "slices"

type Player struct {
	ships   ShipCoordinates
	id      string
	guesses []int
}

func (p *Player) validateGuess(guess int) error {
	if guess < 0 || guess >= 100 {
		return &malformedRequest{
			status: 400, // Bad Request
			msg:    "This guess is out of bounds from the grid",
		}
	}
	if slices.Contains(p.guesses, guess) {
		return &malformedRequest{
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

	p.guesses = append(p.guesses, guess)

	return nil
}

func (p *Player) getHitsAndMisses(other *Player) ([]int, []int) {
	hitMap := make(map[int]bool)
	var hits, misses []int

	ships := other.ships.getFlattenedCoords()

	for _, ship := range ships {
		hitMap[ship] = true
	}

	for _, guess := range p.guesses {
		if hitMap[guess] {
			hits = append(hits, guess)
		} else {
			misses = append(misses, guess)
		}
	}
	return hits, misses
}
