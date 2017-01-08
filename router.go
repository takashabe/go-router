package router

import (
	"fmt"
	"net/http"
)

/*
Design

* simple
* micro wrapper on net/http
* choosable HTTP method. GET, POST...
* parsable URL parameter. "/user/:id"
	* mapping method args. "/user/:user_id/asset/:asset_id" => func(..., userId int, assetId int)
		1. define => router.HandleFunc("/user/:id", userHandler)
		2. parse url path, and extra url params(:id)
		3. call userHandler to with params

*/

type Routing interface {
	Lookup(string) (RouteHandler, error)
	Construct([]*Route) error
}

type Router struct {
	NotFoundHandler http.Handler
	routes          []*Route
	routing         Routing
}

func New() *Router {
	return &Router{
		NotFoundHandler: http.NotFoundHandler(),
		routing:         &Trie{},
	}
}

func (r *Router) SetRouting(routing Routing) {
	r.routing = routing
}

// Building routing tree
func (r *Router) Construct() error {
	return r.routing.Construct(r.routes)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	fmt.Fprintf(w, "Hello, world! %s", path)
}

func (r *Router) HandleFunc(path string, h RouteHandler) *Route {
	return r.AddRoute().HandleFunc(path, h)
}

func (r *Router) AddRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}
