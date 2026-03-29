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

	app.logger.Info("received ship placement request", "data", data)

	gameID := uuid.New()
	app.setCookie(w, gameID)
	app.logger.Info("Get to here", "gameID", gameID.String())
	app.mu.Lock()
	app.logger.Info("Locked")
	app.games[gameID.String()] = data
	app.logger.Info("Unlock")
	app.mu.Unlock()
	app.logger.Info("Get to there")
	app.logger.Info("Game created", "id", gameID)

	response := map[string]string{"gameID": gameID.String()}
	if err := app.writeJSON(w, http.StatusAccepted, response, nil); err != nil {
		http.Error(w, "Issue with sending JSON Response", http.StatusInternalServerError)
	}
}

func (app *application) getGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := app.readCookie(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	app.logger.Info("This should have the game id from a cookie", "gameID", gameID)

	app.mu.Lock()
	gameData, exists := app.games[gameID]
	app.logger.Info("This should have game data", "gameData", gameData)
	app.mu.Unlock()

	if !exists {
		http.Error(w, "Game ID not found", http.StatusNotFound)
		return
	}

	// Return the ships so the "Play" component can render them
	if err := app.writeJSON(w, http.StatusOK, gameData, nil); err != nil {
		http.Error(w, "Issue with sending JSON Response", http.StatusInternalServerError)
		return
	}
}

func (app *application) postGameHandler(w http.ResponseWriter, r *http.Request) {
	
}

// This endpoint only exists to verify cookies are setting properly.  In actual use setCookie() would be called by another handler
// when deemed necessary
// func (app *application) setCookieHandler(w http.ResponseWriter, r *http.Request) {
// 	cookieVal := app.setCookie(w)
// 	app.mu.Lock()
// 	app.sessions[cookieVal] = Session{Active: true, CreatedAt: time.Now()} // user added to Session map
// 	app.mu.Unlock()
// 	w.Write([]byte(cookieVal))
// }

// This endpoint only exists to verify cookies can be read.  In actual use readCookie() will be called by other handlers
// like recieveAttack or createGame to tell game logic which user/game to operate on
func (app *application) getCookieHandler(w http.ResponseWriter, r *http.Request) {
	cookieVal, err := app.readCookie(r)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(cookieVal))
}

// Again, testing handler: verifies sessions are being created in the setCookieHandler
func (app *application) getActiveUsers(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	err := app.writeJSON(w, http.StatusOK, app.sessions, nil)
	if err != nil {
		app.logger.Error("json encoding failed", "error", err.Error())
		http.Error(w, "server error", http.StatusInternalServerError)
	}
}
