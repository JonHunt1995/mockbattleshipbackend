package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func createTestGame(p1ID, p2ID string) *Game {
	p1 := NewPlayer(ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}, p1ID)

	p2 := NewPlayer(ShipCoordinates{
		Carrier:    []int{0, 1, 2, 3, 4},
		Battleship: []int{10, 20, 30, 40},
		Cruiser:    []int{50, 51, 52},
		Submarine:  []int{60, 61, 62},
		Destroyer:  []int{70, 71},
	}, p2ID)

	return NewGame([]*Player{p1, p2})
}

func TestConcurrentGame_PlayTurnAndState(t *testing.T) {
	p1ID := "player-1"
	p2ID := "player-2"
	game := createTestGame(p1ID, p2ID)

	var wg sync.WaitGroup
	iterations := 20

	// Goroutine reading game state concurrently
	stopReader := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopReader:
				return
			default:
				game.mu.Lock()
				_, _ = game.getGameState(p1ID)
				game.mu.Unlock()
			}
		}
	}()

	// Simulate alternating turns between players
	for i := 0; i < iterations; i++ {
		playerID := p1ID
		if i%2 == 1 {
			playerID = p2ID
		}
		guess := (i * 2) % 100

		game.mu.Lock()
		_ = game.playTurn(playerID, guess)
		game.mu.Unlock()
	}

	close(stopReader)
	wg.Wait()
}

func TestConcurrentApp_GetActiveGamesRace(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := &application{
		logger: logger,
		games:  map[string]*Game{},
	}

	gameID := "test-game-1"
	game := createTestGame("player-1", "player-2")
	app.games[gameID] = game

	var readersWg sync.WaitGroup
	var writersWg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Goroutines performing moves on the game
	writersWg.Add(1)
	go func() {
		defer writersWg.Done()
		guess := 0
		for {
			select {
			case <-stopSignal:
				return
			default:
				game.mu.Lock()
				playerID := "player-1"
				if game.Turn%2 == 0 {
					playerID = "player-2"
				}
				_ = game.playTurn(playerID, guess%100)
				guess++
				game.mu.Unlock()
			}
		}
	}()

	// Goroutines hitting getActiveGames (which calls json.Marshal on app.games)
	for i := 0; i < 5; i++ {
		readersWg.Add(1)
		go func() {
			defer readersWg.Done()
			for j := 0; j < 50; j++ {
				req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
				rec := httptest.NewRecorder()
				app.getActiveGames(rec, req)
			}
		}()
	}

	readersWg.Wait()
	close(stopSignal)
	writersWg.Wait()
}

func BenchmarkGame_PlayTurn(b *testing.B) {
	game := createTestGame("player-1", "player-2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerID := "player-1"
		if i%2 == 1 {
			playerID = "player-2"
		}
		guess := (i % 99)
		game.mu.Lock()
		_ = game.playTurn(playerID, guess)
		game.mu.Unlock()
	}
}

func BenchmarkGame_GetGameState(b *testing.B) {
	game := createTestGame("player-1", "player-2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.mu.Lock()
		_, _ = game.getGameState("player-1")
		game.mu.Unlock()
	}
}
