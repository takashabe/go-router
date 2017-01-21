package router

import (
	"log"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
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

var (
	ErrNotFoundHandler = errors.New("not found matched handler")
)

type Routing interface {
	Lookup(method, path string) (HandlerData, error)
	Construct(routes []*Route) error
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

func NewRouter() *Router {
	return &Router{
		NotFoundHandler: http.NotFoundHandler(),
		routing:         NewTrie(),
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
	hd, err := r.routing.Lookup(req.URL.Path, req.Method)
	if err != nil {
		log.Printf("%+v\n", err)
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}

	ref := reflect.ValueOf(hd.handler)
	if ref.Kind() != reflect.Func {
		// TODO: error
		log.Println("Handler is must be Func. but handler kind:%v, path:%v", ref.Kind(), req.URL.Path)
	}
	args, err := parseParams(w, req, hd)
	if err != nil {
		log.Printf("%+v\n", err)
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}
	ref.Call(args)
}

// params convert to []reflect.Value
func parseParams(w http.ResponseWriter, req *http.Request, hd HandlerData) ([]reflect.Value, error) {
	// static args
	args := []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(req),
	}
	// dynamic args
	for _, v := range hd.params {
		args = append(args, reflect.ValueOf(v))
	}
	if reflect.TypeOf(hd.handler).NumIn() != len(args) {
		log.Printf("NumIn()=%v, len(args)=%d\n", reflect.TypeOf(hd.handler).NumIn(), len(args))
		return nil, errors.Wrapf(ErrNotFoundHandler, "path=%s, handler=%v", req.URL.Path, hd.handler)
	}
	return args, nil
}

func (r *Router) Get(path string, h baseHandler) *Route {
	return r.AddRoute().HandleFunc(path, "GET", h)
}

func (r *Router) Post(path string, h baseHandler) *Route {
	return r.AddRoute().HandleFunc(path, "POST", h)
}

func (r *Router) HandleFunc(method, path string, h baseHandler) *Route {
	return r.AddRoute().HandleFunc(method, path, h)
}

func (r *Router) AddRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}
