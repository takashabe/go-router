package router

import (
	"log"
	"net/http"
	"reflect"
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

const paramPrefix = ":"

type Routing interface {
	Lookup(string) (HandlerData, error)
	Construct([]*Route) error
}

type HandlerData struct {
	handler baseHandler
	params  []interface{}
}

type baseHandler interface{}

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
	handler, err := r.routing.Lookup(req.URL.Path)
	log.Println("handler=%v", handler)
	if err != nil {
		log.Printf(err.Error())
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}

	ref := reflect.ValueOf(handler)
	if ref.Kind() != reflect.Func {
		log.Println("Handler is must be Func. but handler kind:%v, path:%v", ref.Kind(), req.URL.Path)
	}
	// static args
	args := []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(req),
	}
	// choose dynamic args
	for i := 2; i < reflect.TypeOf(handler).NumIn(); i++ {

	}
	ref.Call(args)
}

func (r *Router) HandleFunc(path string, h baseHandler) *Route {
	return r.AddRoute().HandleFunc(path, h)
}

func (r *Router) AddRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}
