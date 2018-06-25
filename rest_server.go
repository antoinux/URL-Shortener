package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

// The RestServer listens and handles requests by updating the
// Hash Manager.
// The requests are correctly processed thanks to its own Router.
type RestServer struct {
	Router  *mux.Router
	Manager HashManager
}

// NewRestServer returns a RestServer that is able to serve on the
// specified routes.
func NewRestServer(routes []Route) RestServer {
	router := mux.NewRouter().StrictSlash(true)
	manager := NewHashManager()
	server := RestServer{
		Router:  router,
		Manager: manager,
	}

	for i, route := range routes {
		// funcName declared inside the loop to properly bind it inside
		// the asynchronous closure.
		funcName := routes[i].HandlerFuncName
		server.Router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reflect.ValueOf(&server).MethodByName(funcName).Call(
					[]reflect.Value{
						reflect.ValueOf(w),
						reflect.ValueOf(r),
					},
				)
			})
	}

	return server
}

// Start makes the RestServer start listening on the specified port number.
func (s *RestServer) Start(port int) {
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), s.Router))
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	if encodeErr := json.NewEncoder(w).
		Encode(map[string]string{"ERROR": err.Error()}); encodeErr != nil {
		panic(encodeErr)
	}
}

// Get retrieves the URL corresponding to the requested hash.
func (s *RestServer) Get(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	url, err := s.Manager.Get(hash)
	if err != nil {
		handleError(w, err)
		return
	}
	writeResponse(w, url)
}

// Add inserts the requested URL and finds an available hash for it.
func (s *RestServer) Add(w http.ResponseWriter, r *http.Request) {
	var err error
	var body []byte
	body, err = ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		handleError(w, err)
		return
	}
	if err = r.Body.Close(); err != nil {
		handleError(w, err)
		return
	}

	var entry map[string]string
	if err = json.Unmarshal(body, &entry); err != nil {
		handleError(w, err)
		return
	}
	url, ok := entry["URL"]
	if !ok {
		handleError(w, fmt.Errorf("No field \"URL\" in request."))
		return
	}

	var hash string
	hash, err = s.Manager.Add(url)
	if err != nil {
		handleError(w, err)
		return
	}
	writeResponse(w, hash)
}

// Delete removes the requested hash and its associated URL.
func (s *RestServer) Delete(w http.ResponseWriter, r *http.Request) {
	var err error

	hash := mux.Vars(r)["hash"]
	err = s.Manager.Delete(hash)
	if err != nil {
		handleError(w, err)
		return
	}
	writeResponse(w, hash)
}
