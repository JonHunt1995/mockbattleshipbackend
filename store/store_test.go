package store

import (
	"context"
	"testing"

	"github.com/JonHunt1995/mockbattleshipbackend/player"
)

func TestMemGameStore_CreateAndGet(t *testing.T) {
	s := NewMemGameStore()
	ctx := context.Background()

	// Game does not exist initially
	_, err := s.Get(ctx, "test-game")
	if err == nil {
		t.Fatalf("expected error getting non-existent game, got nil")
	}

	// Create game
	if err := s.Create(ctx, "test-game"); err != nil {
		t.Fatalf("unexpected error creating game: %v", err)
	}

	// Get created game
	g, err := s.Get(ctx, "test-game")
	if err != nil {
		t.Fatalf("unexpected error getting created game: %v", err)
	}
	if g == nil {
		t.Fatalf("expected non-nil game")
	}

	// Create duplicate game
	if err := s.Create(ctx, "test-game"); err == nil {
		t.Fatalf("expected error creating existing game, got nil")
	}
}

func TestMemGameStore_ListJoinable(t *testing.T) {
	s := NewMemGameStore()
	ctx := context.Background()

	// Empty store returns empty slice and nil error
	lobby, err := s.ListJoinable(ctx)
	if err != nil {
		t.Fatalf("unexpected error when no games available: %v", err)
	}
	if len(lobby) != 0 {
		t.Fatalf("expected 0 games in lobby, got %d", len(lobby))
	}

	// Create a game (0 players -> joinable)
	if err := s.Create(ctx, "game-1"); err != nil {
		t.Fatalf("unexpected error creating game-1: %v", err)
	}

	lobby, err = s.ListJoinable(ctx)
	if err != nil {
		t.Fatalf("unexpected error listing joinable games: %v", err)
	}
	if len(lobby) != 1 {
		t.Fatalf("expected 1 game in lobby, got %d", len(lobby))
	}
	if lobby[0].ID != "game-1" || lobby[0].PlayerCount != 0 {
		t.Fatalf("unexpected lobby entry: %+v", lobby[0])
	}

	// Add 2 players to make it unjoinable
	g, err := s.Get(ctx, "game-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	g.Mu.Lock()
	g.Players = append(g.Players, &player.Player{Id: "p1"}, &player.Player{Id: "p2"})
	g.Mu.Unlock()

	lobby, err = s.ListJoinable(ctx)
	if err != nil {
		t.Fatalf("unexpected error when all games are full: %v", err)
	}
	if len(lobby) != 0 {
		t.Fatalf("expected 0 games when all full, got %d", len(lobby))
	}

	// Create a second joinable game
	if err := s.Create(ctx, "game-2"); err != nil {
		t.Fatalf("unexpected error creating game-2: %v", err)
	}

	lobby, err = s.ListJoinable(ctx)
	if err != nil {
		t.Fatalf("unexpected error listing joinable games: %v", err)
	}
	if len(lobby) != 1 || lobby[0].ID != "game-2" {
		t.Fatalf("expected game-2 in lobby, got %+v", lobby)
	}
}
