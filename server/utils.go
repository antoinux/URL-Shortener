package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type Job struct {
	ExecFunc     string
	UrlArg       map[string]string
	Body         []byte
	ResponseChan chan Result
}

type Result struct {
	Response string
	Err      error
}

// writeResponse sends the request's response as json.
// Panics if response can't be encoded as json.
func writeResponse(w http.ResponseWriter, response string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"SUCCESS": response}); err != nil {
		panic(err)
	}
}

// handleError sends the error encountered during the request's processing.
// Panics if err.Error() can't be encoded as json.
func handleError(w http.ResponseWriter, err error) {
	log.Printf("Error on REST server: %v", err)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	if encodeErr := json.NewEncoder(w).
		Encode(map[string]string{"ERROR": err.Error()}); encodeErr != nil {
		panic(encodeErr)
	}
}
