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
type URLServer struct {
	frontRouter *mux.Router
	restRouter  *mux.Router
	manager     *HashManager
}

// NewRestServer returns a RestServer that is able to serve on the
// specified routes.
func NewURLServer(routes []Route) *URLServer {
	log.Println("Building server")
	frontRouter := mux.NewRouter().StrictSlash(true)
	restRouter := mux.NewRouter().StrictSlash(true)
	manager := NewHashManager()
	server := &URLServer{
		frontRouter: frontRouter,
		restRouter:  restRouter,
		manager:     manager,
	}

	log.Println("Routes for REST server:")
	for i, route := range routes {
		// funcName declared inside the loop to properly bind it inside
		// the asynchronous closure.
		log.Printf("Use Route %s with Method %s on Path %s", route.Name, route.Method, route.Pattern)
		funcName := routes[i].HandlerFuncName
		restRouter.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reflect.ValueOf(server).MethodByName(funcName).Call(
					[]reflect.Value{
						reflect.ValueOf(w),
						reflect.ValueOf(r),
					},
				)
			})
	}

	log.Println("Routes for front server:")
	frontRouter.
		Methods("GET").
		Path("/{hash}").
		Name("Redirect").
		HandlerFunc(server.Redirect)

	return server
}

// Start makes the RestServer start listening on the specified port number.
func (s *URLServer) Start(frontPort int, restPort int) {
	log.Printf("Front server listening on port %v\n", frontPort)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", frontPort), s.frontRouter))
	}()
	log.Printf("REST server listening on port %v\n", restPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", restPort), s.restRouter))
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

func (s *URLServer) Redirect(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	log.Println("Recieved redirect request on front server ; hash = %s", hash)
	url, err := s.manager.Get(hash)
	if err != nil {
		handleError(w, err)
		return
	}
	log.Println("Redirecting hash %s to %s", hash, url)
	http.Redirect(w, r, url, 301)
}

// Get retrieves the URL corresponding to the requested hash.
func (s *URLServer) Get(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	log.Println("Recieved Get request on REST server ; hash = %s", hash)
	url, err := s.manager.Get(hash)
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Responding to Get request on REST server ; %s -> %s", hash, url)
	writeResponse(w, url)
}

// Add inserts the requested URL and finds an available hash for it.
func (s *URLServer) Add(w http.ResponseWriter, r *http.Request) {
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

	log.Println("Recieved Add request on REST server ; url %s", url)
	var hash string
	hash, err = s.manager.Add(url)
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Responding to Add request on REST server ; %s -> %s", hash, url)
	writeResponse(w, hash)
}

// Delete removes the requested hash and its associated URL.
func (s *URLServer) Delete(w http.ResponseWriter, r *http.Request) {
	var err error
	hash := mux.Vars(r)["hash"]
	log.Printf("Recieved Delete request on REST server ; hash = %s", hash)
	err = s.manager.Delete(hash)
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Responding to Delete request on REST server ; hash = %s deleted", hash)
	writeResponse(w, hash)
}
