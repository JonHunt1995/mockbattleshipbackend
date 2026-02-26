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
	err := app.readJSON(w, r, &data)
	app.handleDecodeError(w, err)

	app.logger.Info("received ship placement request", "data", data)
}
