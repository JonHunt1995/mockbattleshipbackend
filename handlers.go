package main

import (
	"fmt"
	"net"
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
		app.notFoundResponse(w, r, err)
		return
	}

	playerID, err := app.findUserSession(r)
	if err != nil {
		playerUUID := uuid.New()
		app.setUserSession(w, playerUUID)
		playerID = playerUUID.String()
	}

	player := NewPlayer(data, playerID)

	game, err := app.getGame(gameID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	game.mu.Lock()
	err = game.addPlayer(player)
	game.mu.Unlock()

	if err != nil {
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
	InviteLink string `json:"link"`
}

func (app *application) createNewGame(w http.ResponseWriter, r *http.Request) {
	hostName := r.Header.Get("X-Original-Host")
	if hostName == "" {
		hostName = r.Host
	}

	domain, _, err := net.SplitHostPort(hostName)
	if err != nil {
		domain = hostName
	}
	gameID := uuid.New()

	if err := app.setGame(gameID.String()); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	inviteURL := fmt.Sprintf("%s/setup/%s", domain, gameID)

	payload := &createGameResponse{InviteLink: inviteURL}

	if err := app.writeJSON(w, http.StatusAccepted, payload, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getGameHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := app.findUserSession(r)
	gameID := r.PathValue("gameID")

	if gameID == "" {
		app.notFoundResponse(w, r, err)
		return
	}

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	game, err := app.getGame(gameID)

	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	game.mu.Lock()
	player, err := game.getPlayer(playerID)
	if err != nil {
		game.mu.Unlock()
		app.badRequestResponse(w, r, err)
		return
	}

	gs, err := game.getGameState(player.Id)
	game.mu.Unlock()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, gs, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

type playerMove struct {
	Guess int `json:"Guess"`
}

func (pm *playerMove) getGuess() int {
	return pm.Guess
}

func (app *application) postGameHandler(w http.ResponseWriter, r *http.Request) {
	// this will be the handler that actually deals with game move logic
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

	playerID, err := app.findUserSession(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	gameID := r.PathValue("gameID")
	if gameID == "" {
		app.notFoundResponse(w, r, err)
		return
	}

	game, err := app.getGame(gameID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	game.mu.Lock()
	err = game.playTurn(playerID, move.getGuess())
	if err != nil {
		game.mu.Unlock()
		app.badRequestResponse(w, r, err)
		return
	}

	gs, err := game.getGameState(playerID)
	game.mu.Unlock()

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
	app.mu.RLock()
	gamePointers := make(map[string]*Game, len(app.games))
	for id, g := range app.games {
		gamePointers[id] = g
	}
	app.mu.RUnlock()

	snapshots := make(map[string]*ClonedGame, len(gamePointers))
	for id, g := range gamePointers {
		snapshots[id] = g.Clone()
	}

	err := app.writeJSON(w, http.StatusOK, snapshots, nil)
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

	numPlayers, err := strconv.Atoi(r.PathValue("numPlayers"))
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
	currentTurn := game.Turn
	playerCount := game.getPlayerCount()
	game.mu.Unlock()

	if turn == currentTurn && numPlayers == playerCount {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Client is not up to date, let's trigger a resync
	w.WriteHeader(http.StatusOK)
}
