package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (app *application) setupGameHandler(w http.ResponseWriter, r *http.Request) {
	var data ShipCoordinates

	if err := app.readJSON(w, r, &data); err != nil {
		app.handleDecodeError(w, err)
		return
	}

	if !data.areValid() {
		http.Error(w, "The uploaded ships are not properly placed", http.StatusBadRequest)
		return
	}

	gameID := "the only game that matters"
	playerID := uuid.New()
	player := NewPlayer(data, playerID.String())

	app.logger.Info("received ship placement request", "data", data)

	app.mu.Lock()
	app.logger.Info("Locked")
	if _, ok := app.games[gameID]; !ok {
		http.Error(w, "game doesn't exist", http.StatusBadRequest)
	}
	if err := app.games[gameID].addPlayer(player); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	app.logger.Info("Unlock")
	app.mu.Unlock()

	app.logger.Info("Get to there")
	app.logger.Info("Game created", "id", gameID)

	app.setCookie(w, playerID, true)
	response := map[string]string{"gameID": gameID}
	if err := app.writeJSON(w, http.StatusAccepted, response, nil); err != nil {
		http.Error(w, "Issue with sending JSON Response", http.StatusInternalServerError)
	}
}

func (app *application) createNewGame(w http.ResponseWriter, r *http.Request) {
	var data Game

	if err := app.readJSON(w, r, &data); err != nil {
		app.handleDecodeError(w, err)
		return
	}
	gameID := uuid.New()
	app.setCookie(w, gameID, false)
}

type LivingShips struct {
	Carrier bool
	Battleship bool
	Cruiser bool
	Submarine bool
	
}
type gameStateResponse struct {
	PlayerShips    []int
	PlayerHits     []int
	OpponentHits   []int
	OpponentMisses []int
}

func (app *application) getGameHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := app.readCookie(r, true)
	gameID := "the only game that matters"

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//app.logger.Info("This should have the game id from a cookie", "gameID", gameID)

	app.mu.Lock()
	gameData, exists := app.games[gameID]
	app.logger.Info("This should have game data", "gameData", gameData)
	app.mu.Unlock()

	if !exists {
		app.notFoundResponse(w, r)
		return
	}

	playerData, err := gameData.getPlayer(playerID)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	opponentData, err := gameData.getOpponent(playerID)
	app.logger.Info("This should be either opponents data or nil", "opponentData", opponentData, "error", err)
	var opponentShips ShipCoordinates

	if err == nil {
		opponentShips = opponentData.Ships
	}

	gs := &gameStateResponse{
		PlayerShips:   playerData.Ships,
		PlayerHits: playerData.,
		PlayerMisses: ,
		OpponentHits: opponentShips,
		OpponentMisses: ,
	}

	app.logger.Info("gameState", "Player", playerData, "Opponent", opponentData)
	// Return the ships so the "Play" component can render them
	if err := app.writeJSON(w, http.StatusOK, gs, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) postGameHandler(w http.ResponseWriter, r *http.Request) {

}

// Again, testing handler: verifies sessions are being created in the setCookieHandler
func (app *application) getActiveGames(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	err := app.writeJSON(w, http.StatusOK, app.games, nil)
	if err != nil {
		app.logger.Error("json encoding failed", "error", err.Error())
		http.Error(w, "server error", http.StatusInternalServerError)
	}
}
