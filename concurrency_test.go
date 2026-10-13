package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/JonHunt1995/mockbattleshipbackend/game"
	"github.com/JonHunt1995/mockbattleshipbackend/player"
	"github.com/JonHunt1995/mockbattleshipbackend/ship"
	"github.com/JonHunt1995/mockbattleshipbackend/store"
)

func createTestGame(p1ID, p2ID string) *game.Game {
	p1 := player.NewPlayer(ship.ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}, p1ID)

	p2 := player.NewPlayer(ship.ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}, p2ID)

	return game.NewGame([]*player.Player{p1, p2})
}

func TestConcurrentGame_PlayTurnAndState(t *testing.T) {
	p1ID := "player-1"
	p2ID := "player-2"
	g := createTestGame(p1ID, p2ID)

	var wg sync.WaitGroup
	iterations := 20

	// Goroutine reading game state concurrently
	stopReader := make(chan struct{})
	wg.Go(func() {
		for {
			select {
			case <-stopReader:
				return
			default:
				g.Mu.Lock()
				_, _ = g.GetGameState(p1ID)
				g.Mu.Unlock()
			}
		}
	})

	// Simulate alternating turns between players
	for i := range iterations {
		playerID := p1ID
		if i%2 == 1 {
			playerID = p2ID
		}
		guess := (i * 2) % 100

		g.Mu.Lock()
		_ = g.PlayTurn(playerID, guess)
		g.Mu.Unlock()
	}

	close(stopReader)
	wg.Wait()
}

func TestConcurrentApp_LobbyRace(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := &application{
		config: config{port: ":4000"},
		logger: logger,
		store:  store.NewMemGameStore(),
	}

	gameID := "test-game-1"
	if err := app.store.Create(context.Background(), gameID); err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	g, err := app.store.Get(context.Background(), gameID)
	if err != nil {
		t.Fatalf("failed to get game: %v", err)
	}

	p1 := player.NewPlayer(ship.ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}, "player-1")

	p2 := player.NewPlayer(ship.ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}, "player-2")

	_ = g.AddPlayer(p1)
	_ = g.AddPlayer(p2)

	var readersWg sync.WaitGroup
	var writersWg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Goroutines performing moves on the game
	writersWg.Go(func() {
		guess := 0
		for {
			select {
			case <-stopSignal:
				return
			default:
				g.Mu.Lock()
				playerID := "player-1"
				if g.Turn%2 == 0 {
					playerID = "player-2"
				}
				_ = g.PlayTurn(playerID, guess%100)
				guess++
				g.Mu.Unlock()
			}
		}
	})

	// Goroutines hitting getLobbyHandler (which calls ListJoinable)
	for range 5 {
		readersWg.Go(func() {
			for range 50 {
				req := httptest.NewRequest(http.MethodGet, "/api/lobby", nil)
				rec := httptest.NewRecorder()
				app.getLobbyHandler(rec, req)
			}
		})
	}

	readersWg.Wait()
	close(stopSignal)
	writersWg.Wait()
}

func BenchmarkGame_PlayTurn(b *testing.B) {
	g := createTestGame("player-1", "player-2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerID := "player-1"
		if i%2 == 1 {
			playerID = "player-2"
		}
		guess := (i % 99)
		g.Mu.Lock()
		_ = g.PlayTurn(playerID, guess)
		g.Mu.Unlock()
	}
}

func BenchmarkGame_GetGameState(b *testing.B) {
	g := createTestGame("player-1", "player-2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Mu.Lock()
		_, _ = g.GetGameState("player-1")
		g.Mu.Unlock()
	}
}
