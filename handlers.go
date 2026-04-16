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

	if err := data.areValid(); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	gameID := r.PathValue("gameID")
	var err error

	if gameID == "" {
		gameID, err = app.readCookie(r, false)
		if err != nil {
			app.notFoundResponse(w, r)
		}
	}

	playerID := uuid.New()
	player := NewPlayer(data, playerID.String())

	app.logger.Info("received ship placement request", "data", data)

	game, err := app.getGame(gameID)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.addPlayer(player); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.setCookie(w, playerID, true)

	response := map[string]string{"gameID": gameID}
	if err := app.writeJSON(w, http.StatusAccepted, response, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createNewGame(w http.ResponseWriter, r *http.Request) {
	gameID := uuid.New()
	playerID := uuid.New()
	app.setCookie(w, gameID, false)
	app.setCookie(w, playerID, true)

	if err := app.setGame(gameID.String()); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

type gameStateResponse struct {
	PlayerShips         []int
	PlayerHits          []int
	PlayerMisses        []int
	PlayerLivingShips   LivingShips
	OpponentHits        []int
	OpponentMisses      []int
	OpponentLivingShips LivingShips
}

func (app *application) getGameHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := app.readCookie(r, true)
	gameID := "the only game that matters"

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()
	gameData, exists := app.games[gameID]
	app.logger.Info("This should have game data", "gameData", gameData)

	if !exists {
		app.notFoundResponse(w, r)
		return
	}

	player, err := gameData.getPlayer(playerID)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	opponent, err := gameData.getOpponent(playerID)
	app.logger.Info("This should be either opponents data or nil", "opponent", opponent, "error", err)
	var opponentHits []int
	var opponentMisses []int
	var opponentLivingShips LivingShips
	var playerHits []int
	var playerMisses []int
	playerLivingShips := &LivingShips{
		Carrier:    true,
		Battleship: true,
		Cruiser:    true,
		Submarine:  true,
		Destroyer:  true,
	}

	if err == nil {
		opponentHits, opponentMisses = opponent.getHitsAndMisses(player)

	}

	gs := &gameStateResponse{
		PlayerShips:         player.Ships.getFlattenedCoords(),
		PlayerHits:          playerHits,
		PlayerMisses:        playerMisses,
		PlayerLivingShips:   *playerLivingShips,
		OpponentHits:        opponentHits,
		OpponentMisses:      opponentMisses,
		OpponentLivingShips: opponentLivingShips,
	}

	app.logger.Info("gameState", "Player", player, "Opponent", opponent)
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
