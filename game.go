package main

import (
	"sync"
)

type Game struct {
	Players []*Player
	Turn    int
	mu      sync.Mutex
}

func NewGame(players []*Player) *Game {
	return &Game{
		Players: players,
		Turn:    1,
		mu:      sync.Mutex{},
	}
}

func (g *Game) getPlayerCount() int {
	return len(g.Players)
}

func (g *Game) addPlayer(p *Player) error {
	if g.getPlayerCount() >= 2 {
		return &badRequest{
			status: 400,
			msg:    "This game already has 2 players",
		}
	}
	g.Players = append(g.Players, p)

	return nil
}

func (g *Game) getPlayerNumber(player_id string) (int, bool) {
	for i, p := range g.Players {
		if p.Id == player_id {
			return i + 1, true
		}
	}

	return 0, false
}

func (g *Game) getPlayer(playerId string) (*Player, error) {
	idx, found := g.getPlayerNumber(playerId)

	if !found {
		return nil, &badRequest{
			status: 400,
			msg:    "Player not found",
		}
	}

	idx -= 1

	return g.Players[idx], nil
}

func (g *Game) getOpponent(playerId string) (*Player, error) {
	other, found := g.getPlayerNumber(playerId)

	if !found {
		return nil, &badRequest{
			status: 400,
			msg:    "Player not found",
		}
	}

	other %= 2

	if len(g.Players) <= 1 || g.Players[other].Id == playerId {
		return nil, &badRequest{
			status: 400,
			msg:    "No opponent found",
		}
	}

	return g.Players[other], nil
}

func (g *Game) validateTurn(player_id string, guess int) error {
	playersTurn := (g.Turn + 1) % 2
	if g.Players[playersTurn].Id == player_id {
		return nil
	}
	return &badRequest{
		status: 400,
		msg:    "It is not your turn",
	}
}

func (g *Game) playTurn(player_id string, guess int) error {
	if err := g.validateTurn(player_id); err != nil {
		return err
	}

	if err := g.Players[g.Turn%2].AddGuess(guess); err != nil {
		return err
	}

	return nil
}
