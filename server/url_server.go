package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"restful/manager"
)

var MAXJOBS int = 4

// The RestServer listens and handles requests by updating the
// Hash Manager.
// The requests are correctly processed thanks to its own Router.
type URLServer struct {
	frontRouter *mux.Router
	restRouter  *mux.Router
	manager     *manager.HashManager
	jobQueue    chan Job
}

func (s *URLServer) worker() {
	// Explanation: worker waits for jobs on the jobQueue.
	// When a job is here, we need to use reflection because the method to call
	// is not static (this is on purpose) since it is a member function.
	// Reflection is quite ugly, in particular for that last interface{}
	// casting line.
	//
	// Moreover, we can't use the http.responseWriter and http.Request inside
	// the worker since they are closed when the parent goroutine ends, so we
	// need to send back the output via another channel.
	// Note: we could use the http objects here, but the parent goroutine would
	// have to wait anyway, so a channel is necessary.
	for {
		job := <-s.jobQueue
		job.ResponseChan <- reflect.ValueOf(s).MethodByName(job.ExecFunc).Call(
			[]reflect.Value{
				reflect.ValueOf(job.Request),
				reflect.ValueOf(job.Body),
			},
		)[0].Interface().(Result)
	}
}

// NewURLServer returns a URLServer that is able to serve two types of
// requests, the RESTful requests on the specified routes, and the
// front end requests that need to be redirected.
func NewURLServer(routes []Route) *URLServer {
	log.Println("Building server")
	frontRouter := mux.NewRouter().StrictSlash(true)
	restRouter := mux.NewRouter().StrictSlash(true)
	manager := manager.NewHashManager()
	server := &URLServer{
		frontRouter: frontRouter,
		restRouter:  restRouter,
		manager:     manager,
		jobQueue:    make(chan Job),
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
				// Read here since the body is closed at the end of this
				// routine.
				body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
				if err != nil {
					handleError(w, err)
					return
				}

				responseChan := make(chan Result)
				server.jobQueue <- Job{
					ExecFunc:     funcName,
					Request:      r,
					Body:         body,
					ResponseChan: responseChan,
				}

				// Write from here since the writer will be closed at the end
				// of this routine.
				result := <-responseChan
				if result.Err != nil {
					handleError(w, result.Err)
				} else {
					writeResponse(w, result.Response)
				}
			})
	}
	log.Println("Routes for front server:")
	log.Printf("Use Route Redirect with Method GET on Path /{hash}.")
	frontRouter.
		Methods("GET").
		Path("/{hash}").
		Name("Redirect").
		HandlerFunc(server.Redirect)

	for i := 0; i < MAXJOBS; i++ {
		go server.worker()
	}

	return server
}

// Start makes the URLServer start listening on the specified port numbers.
func (s *URLServer) Start(frontPort int, restPort int) {
	log.Printf("Front server listening on port %v\n", frontPort)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", frontPort), s.frontRouter))
	}()
	log.Printf("REST server listening on port %v\n", restPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", restPort), s.restRouter))
}

// Redirect finds the URL corresponding to the given hash and issues
// an HTTP redirect to it.
func (s *URLServer) Redirect(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	log.Printf("Recieved redirect request on front server ; hash = %s\n", hash)
	url, err := s.manager.Get(hash)
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Redirecting hash %s to %s\n", hash, url)
	http.Redirect(w, r, url, 301)
}

// Get retrieves the URL corresponding to the requested hash.
func (s *URLServer) Get(r *http.Request, body []byte) Result {
	hash := mux.Vars(r)["hash"]
	log.Printf("Recieved Get request on REST server ; hash = %s\n", hash)
	url, err := s.manager.Get(hash)
	if err != nil {
		return Result{Err: err}
	}
	log.Printf("Responding to Get request on REST server ; %s -> %s", hash, url)
	return Result{Response: url}
}

// Add inserts the requested URL and finds an available hash for it.
func (s *URLServer) Add(r *http.Request, body []byte) Result {
	var err error
	var entry map[string]string

	if err = json.Unmarshal(body, &entry); err != nil {
		return Result{Err: err}
	}
	url, ok := entry["URL"]
	if !ok {
		return Result{Err: fmt.Errorf("No field \"URL\" in request.")}
	}

	log.Printf("Recieved Add request on REST server ; url %s\n", url)
	var hash string
	hash, err = s.manager.Add(url)
	if err != nil {
		return Result{Err: err}
	}
	log.Printf("Responding to Add request on REST server ; %s -> %s", hash, url)
	return Result{Response: hash}
}

// Delete removes the requested hash and its associated URL.
func (s *URLServer) Delete(r *http.Request, body []byte) Result {
	var err error
	hash := mux.Vars(r)["hash"]
	log.Printf("Recieved Delete request on REST server ; hash = %s", hash)
	err = s.manager.Delete(hash)
	if err != nil {
		return Result{Err: err}
	}
	log.Printf("Responding to Delete request on REST server ; hash = %s deleted", hash)
	return Result{Response: hash}
}
