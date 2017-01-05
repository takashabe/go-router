package router

import "net/http"

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
	Lookup(string) RouteHandler
	Construct([]*Route)
}

type Router struct {
	NotFoundHandler http.Handler
	routes          []*Route
}

func NewRouter() *Router {
	return &Router{
		NotFoundHandler: http.NotFoundHandler(),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}

func (r *Router) HandleFunc(path string, h RouteHandler) *Route {
	return r.AddRoute().HandleFunc(path, h)
}

func (r *Router) AddRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}
