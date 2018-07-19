package server

// Route contains the structure of an endpoint.
type Route struct {
	Name            string
	Method          string // "GET", "POST",...
	Pattern         string // "/",...
	HandlerFuncName string // Method of a RestServer object
}

// The liste of routes the API supports.
// To add a new endpoint, add a new route here and a new function in
// rest_server.go to handle the requests if necessary.
var Routes = []Route{
	Route{
		Name:            "Get",
		Method:          "GET",
		Pattern:         "/{hash}",
		HandlerFuncName: "Get",
	},
	Route{
		Name:            "Add",
		Method:          "POST",
		Pattern:         "/",
		HandlerFuncName: "Add",
	},
	Route{
		Name:            "Delete",
		Method:          "DELETE",
		Pattern:         "/{hash}",
		HandlerFuncName: "Delete",
	},
}
