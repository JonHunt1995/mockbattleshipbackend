package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Heavily influenced by https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
// This could have been like 5 lines if I didn't want to have helpful errors, but I suffered so
// that the frontend devs can have helpful errors.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	ct := r.Header.Get("Content-Type")

	if ct != "" {
		mediatype := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediatype != "application/json" {
			msg := "Content-Type header is not application/json"
			return &badRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}
	// Checking to see if HTTP Response is too large
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	// Error out here if JSON fields don't match destination struct
	dec.DisallowUnknownFields()

	err := dec.Decode(dest)
	// This section is to respond with helpful messages to the frontend
	// about why a request is malformed and how to properly fix it
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError
		var msg string
		status := http.StatusBadRequest

		switch {
		case errors.As(err, &syntaxError):
			msg = fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg = "Request body contains badly-formed JSON"

		case errors.As(err, &unmarshalTypeError):
			msg = fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg = fmt.Sprintf("Request body contains unknown field %s", fieldName)

		case errors.Is(err, io.EOF):
			msg = "Request body must not be empty"

		case errors.As(err, &maxBytesError):
			msg = fmt.Sprintf("Request body must not be larger than %d bytes", maxBytesError.Limit)
			status = http.StatusRequestEntityTooLarge

		default:
			return err
		}

		return &badRequest{
			status: status,
			msg:    msg,
		}
	}

	// Ensure Only Single JSON Object
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		return &badRequest{
			status: http.StatusBadRequest,
			msg:    msg,
		}
	}

	// Success!
	return nil
}

// Using the error handling logic from same blog post and wrapping it into
// a function for sanity.
func (app *application) handleDecodeError(w http.ResponseWriter, err error) {
	var mr *badRequest
	payload := make(map[string]string)

	if errors.As(err, &mr) {
		payload["error"] = mr.msg

		if writeErr := app.writeJSON(w, mr.status, payload, nil); writeErr != nil {
			app.logger.Error("failed to write client error response", "error", writeErr)
		}
		return
	}

	app.logger.Error("unexpected error decoding json", "error", err)
	payload["error"] = http.StatusText(http.StatusInternalServerError)
	if writeErr := app.writeJSON(w, http.StatusInternalServerError, payload, nil); writeErr != nil {
		app.logger.Error("failed to write catch all client error response", "error", writeErr)
	}
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	payload, err := json.Marshal(data)

	if err != nil {
		return err
	}

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(payload)
	// Success!
	return nil
}

// Read cookie value or return an error
func (app *application) readCookie(r *http.Request, isUser bool) (string, error) {
	name := "game"
	if isUser {
		name = "user"
	}

	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", err
		}
		app.logger.Error("Error parsing cookie: ", "error", err.Error())
		return "", err
	}

	return cookie.Value, nil
}

// Write cookie and return the cookie value
func (app *application) setCookie(w http.ResponseWriter, id uuid.UUID, isUser bool) string {
	name := "game"
	if isUser {
		name = "user"
	}
	cookie := http.Cookie{
		Name:     name,
		Value:    id.String(),
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	return id.String()
}
