package main

import (
	"net/http"
	"time"
)

type ShipPlacementRequest struct {
	Carrier    []int `json:"Carrier"`
	Battleship []int `json:"Battleship"`
	Cruiser    []int `json:"Cruiser"`
	Submarine  []int `json:"Submarine"`
	Destroyer  []int `json:"Destroyer"`
}

func (app *application) setupGameHandler(w http.ResponseWriter, r *http.Request) {
	var data ShipPlacementRequest
	if err := app.readJSON(w, r, &data); err != nil {
		app.handleDecodeError(w, err)
		return
	}

	app.logger.Info("received ship placement request", "data", data)
}

// This endpoint only exists to verify cookies are setting properly.  In actual use setCookie() would be called by another handler
// when deemed necessary
func (app *application) setCookieHandler(w http.ResponseWriter, r *http.Request) {
	cookieVal := app.setCookie(w)
	app.mu.Lock()
	app.sessions[cookieVal] = Session{Active: true, CreatedAt: time.Now()} // user added to Session map
	app.mu.Unlock()
	w.Write([]byte(cookieVal))
}

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
