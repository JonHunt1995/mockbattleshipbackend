package main

import "net/http"

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

func (app *application) setCookieHandler(w http.ResponseWriter, r *http.Request) {
	cookieVal := app.setCookie(w)

	w.Write([]byte(cookieVal))
}

func (app *application) getCookieHandler(w http.ResponseWriter, r *http.Request) {
	cookieVal, err := app.readCookie(r)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(cookieVal))
}
