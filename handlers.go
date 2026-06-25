package main

import (
	"fmt"
	"net/http"
	"strconv"

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
			app.notFoundResponse(w, r, err)
			return
		}
	}

	playerID, err := app.readCookie(r, true)
	if err != nil {
		playerUUID := uuid.New()
		app.setCookie(w, playerUUID, true)
		playerID = playerUUID.String()
	}

	player := NewPlayer(data, playerID)

	app.logger.Info("received ship placement request", "data", data)

	game, err := app.getGame(gameID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if err := game.addPlayer(player); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	response := map[string]string{"gameID": gameID}
	if err := app.writeJSON(w, http.StatusAccepted, response, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

type createGameResponse struct {
	InviteLink string
}

func (app *application) createNewGame(w http.ResponseWriter, r *http.Request) {
	// will have to update this in prod from localhost
	link := "http://localhost:5173"
	gameID := uuid.New()
	playerID := uuid.New()
	app.setCookie(w, gameID, false)
	app.setCookie(w, playerID, true)

	if err := app.setGame(gameID.String()); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	inviteURL := fmt.Sprintf("%s/setup/%s", link, gameID)

	payload := &createGameResponse{InviteLink: inviteURL}

	if err := app.writeJSON(w, http.StatusAccepted, payload, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getGameHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := app.readCookie(r, true)
	gameID := r.PathValue("gameID")

	if gameID == "" {
		gameID, err = app.readCookie(r, false)
		if err != nil {
			app.notFoundResponse(w, r, err)
			return
		}
	}

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	game, err := app.getGame(gameID)
	app.logger.Info("This should have game data", "game", game)

	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	player, err := game.getPlayer(playerID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	opponent, err := game.getOpponent(playerID)
	app.logger.Info("This should be either opponents data or nil", "opponent", opponent, "error", err)

	gs, err := game.getGameState(player.Id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.logger.Info("gameState", "Player", player, "Opponent", opponent)

	if err := app.writeJSON(w, http.StatusOK, gs, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

type playerMove struct {
	Guess int `json: "Guess"`
}

func (pm *playerMove) getGuess() int {
	return pm.Guess
}

func (app *application) postGameHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: this will be the handler that actually deals with game move logic
	// by receiving moves from the frontend.
	// the backend will first:
	// - check if move is valid
	// 	- if move isn't valid, send back helpful errors to client and keep game state
	// 		the same
	// 	- otherwise, move is fine and we'll continue moving the chain
	// - apply move
	// - send back updated game state response to respective clients
	// 	- will this pub/sub or how will this work? We could tell the FE to do a PRG
	var move playerMove

	err := app.readJSON(w, r, &move)
	if err != nil {
		app.handleDecodeError(w, err)
		return
	}

	playerID, err := app.readCookie(r, true)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	gameID := r.PathValue("gameID")
	if gameID == "" {
		gameID, err = app.readCookie(r, false)
		if err != nil {
			app.notFoundResponse(w, r, err)
			return
		}
	}

	game, err := app.getGame(gameID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	err = game.playTurn(playerID, move.getGuess())
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	gs, err := game.getGameState(playerID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, gs, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getActiveGames(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	err := app.writeJSON(w, http.StatusOK, app.games, nil)
	if err != nil {
		app.logger.Error("json encoding failed", "error", err.Error())
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) pollHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("gameID")
	if gameID == "" {
		app.notFoundResponse(w, r, nil)
		return
	}

	turn, err := strconv.Atoi(r.PathValue("turn"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}


	game, err := app.getGame(gameID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if turn == game.Turn {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Client is not up to date, let's trigger a resync
	w.WriteHeader(http.StatusOK)
}
