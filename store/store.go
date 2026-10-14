package store

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/JonHunt1995/mockbattleshipbackend/game"
	"github.com/JonHunt1995/mockbattleshipbackend/player"
)

type LobbyGame struct {
	ID          string `json:"id"`
	PlayerCount int    `json:"playerCount"`
}

type GameStore interface {
	Get(ctx context.Context, id string) (*game.Game, error)
	Create(ctx context.Context, id string) error
	ListJoinable(ctx context.Context) ([]LobbyGame, error)
}

type MemGameStore struct {
	games map[string]*game.Game
	mu    sync.RWMutex
}

func NewMemGameStore() *MemGameStore {
	return &MemGameStore{
		games: make(map[string]*game.Game),
	}
}

func (s *MemGameStore) Get(ctx context.Context, id string) (*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, exists := s.games[id]
	if !exists {
		message := fmt.Sprintf("Game %v doesn't exist", id)
		return nil, errors.New(message)
	}

	return g, nil
}

func (s *MemGameStore) Create(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[id]; exists {
		message := fmt.Sprintf("Game %v already exists", id)
		return errors.New(message)
	}

	s.games[id] = game.NewGame([]*player.Player{})

	return nil
}

func (s *MemGameStore) ListJoinable(ctx context.Context) ([]LobbyGame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]LobbyGame, 0)

	for id, g := range s.games {
		g.Mu.Lock()
		if g.GetPlayerCount() <= 1 {
			result = append(result, LobbyGame{ID: id, PlayerCount: g.GetPlayerCount()})
		}
		g.Mu.Unlock()
	}

	return result, nil
}
