package main

import (
	"net/http"
	"time"
)

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		app.logger.Info("request processed",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"duration", time.Since(start),
		)
	})
}

func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// Build the chain from right to left
		// This ensures execution happens left to right
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
