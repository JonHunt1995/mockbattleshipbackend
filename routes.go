package main

import (
	"net/http"

	"github.com/rs/cors"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/setup/{gameID}", app.setupGameHandler)
	mux.HandleFunc("GET /api/play", app.getGameHandler)
	mux.HandleFunc("POST /api/play", app.postGameHandler)
	mux.HandleFunc("GET /api/games", app.getActiveGames)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	return app.logRequest(c.Handler(mux))
}
