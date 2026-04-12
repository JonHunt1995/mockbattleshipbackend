package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

type badRequest struct {
	status int
	msg    string
}

func (mr *badRequest) Error() string {
	return mr.msg
}

type envelope struct {
	Error any `json:"error"`
}

// Heavily influenced by Alex Edwards' book
func (app *application) logError(r *http.Request, err error) {
	app.logger.Error(err.Error(),
		slog.Group("request",
			"method", r.Method,
			"url", r.URL.String(),
		),
	)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{Error: message}
	err := app.writeJSON(w, status, env, nil)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}
