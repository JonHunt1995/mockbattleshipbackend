package main

import (
	"net/http"

	"github.com/rs/cors"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/setup", app.setupGameHandler)
	mux.HandleFunc("GET /api/play", app.getGameHandler)
	mux.HandleFunc("GET /api/get-cookie", app.getCookieHandler)
	//mux.HandleFunc("GET /api/set-cookie", app.setCookieHandler)
	mux.HandleFunc("GET /api/users", app.getActiveUsers)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	return app.logRequest(c.Handler(mux))
}
