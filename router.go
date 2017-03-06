package router

import (
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	// TokenParam represents parameter in the URL path
	// e.g. "/user/:id"
	TokenParam = ":"

	// TokenParam represents wildcard in the URL path
	// e.g. "/static/*filepath
	TokenWildcard = "*"
)

var (
	ErrNotFoundHandler = errors.New("not found matched handler")
	ErrInvalidHandler  = errors.New("invalid handler")
	ErrInvalidParam    = errors.New("invalid param")
)

type Routing interface {
	Lookup(method, path string) (HandlerData, error)
	Insert(method, path string, handler baseHandler) error
}

type HandlerData struct {
	handler baseHandler
	params  []interface{}
}

type baseHandler interface{}

type Router struct {
	NotFoundHandler http.Handler
	Routing         Routing
	routes          []*Route
	OutLog          *log.Logger
	ErrLog          *log.Logger
}

func NewRouter() *Router {
	return &Router{
		NotFoundHandler: http.NotFoundHandler(),
		Routing:         NewTrie(),
		OutLog:          log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		ErrLog:          log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (r *Router) accessLogf(format string, args ...interface{}) {
	if env := os.Getenv("GO_ROUTER_ENABLE_LOGGING"); len(env) != 0 {
		r.OutLog.Printf("%s\n", args)
	}
}

func (r *Router) errorLogf(format string, args ...interface{}) {
	if env := os.Getenv("GO_ROUTER_ENABLE_LOGGING"); len(env) != 0 {
		r.ErrLog.Printf("[error] %s\n", args)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hd, err := r.Routing.Lookup(req.URL.Path, req.Method)
	if err != nil {
		r.errorLogf("not found path: %s. %#v", req.URL.Path, err)
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}

	err = r.callHandler(w, req, hd)
	if err != nil {
		r.errorLogf("failed call handler. %#v", err)
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}
}

func (r *Router) callHandler(w http.ResponseWriter, req *http.Request, hd HandlerData) error {
	ref := reflect.ValueOf(hd.handler)
	if ref.Kind() != reflect.Func {
		return errors.Wrapf(ErrInvalidHandler, "handler is must be Func. got:%v", ref.Kind())
	}
	args, err := r.parseParams(w, req, hd)
	if err != nil {
		return errors.Wrapf(err, "failed parsed params")
	}

	r.accessLogf("%s - - [%s] \"%s %s %s\"", req.RemoteAddr, time.Now().Format(time.RFC822Z), req.Method, req.URL.Path, req.Proto)
	ref.Call(args)
	return nil
}

// params convert to []reflect.Value
func (r *Router) parseParams(w http.ResponseWriter, req *http.Request, hd HandlerData) ([]reflect.Value, error) {
	// static args means http.ResponseWriter and *http.Request
	numStaticArgs := 2

	ref := reflect.TypeOf(hd.handler)
	if ref.NumIn() != len(hd.params)+numStaticArgs {
		return nil, errors.Wrapf(ErrNotFoundHandler, "path=%s, handler=%v", req.URL.Path, hd.handler)
	}

	// static args
	args := []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(req),
	}
	// dynamic args
	for i := numStaticArgs; i < ref.NumIn(); i++ {
		switch ref.In(i).Kind() {
		case reflect.Int:
			p, err := strconv.Atoi(hd.params[i-numStaticArgs].(string))
			if err != nil {
				return nil, errors.Wrapf(ErrInvalidParam, "path=%s, error=%s", req.URL.Path, err)
			}
			args = append(args, reflect.ValueOf(p))
		default:
			args = append(args, reflect.ValueOf(hd.params[i-numStaticArgs]))
		}
	}
	return args, nil
}

func (r *Router) Get(path string, h baseHandler) *Route     { return r.HandleFunc("GET", path, h) }
func (r *Router) Head(path string, h baseHandler) *Route    { return r.HandleFunc("HEAD", path, h) }
func (r *Router) Post(path string, h baseHandler) *Route    { return r.HandleFunc("POST", path, h) }
func (r *Router) Put(path string, h baseHandler) *Route     { return r.HandleFunc("PUT", path, h) }
func (r *Router) Patch(path string, h baseHandler) *Route   { return r.HandleFunc("PATCH", path, h) }
func (r *Router) Delete(path string, h baseHandler) *Route  { return r.HandleFunc("DELETE", path, h) }
func (r *Router) Options(path string, h baseHandler) *Route { return r.HandleFunc("OPTIONS", path, h) }

func (r *Router) HandleFunc(method, path string, h baseHandler) *Route {
	route := r.AddRoute().HandleFunc(method, path, h)
	err := r.Routing.Insert(route.method, route.path, route.handler)
	if err != nil {
		r.errorLogf("failed registered path. path=%s, error=%v", path, err)
	}
	return route
}

func (r *Router) ServeFile(path string, root http.FileSystem) {
	fs := http.FileServer(root)
	r.Get(path, func(w http.ResponseWriter, req *http.Request, suffixPath string) {
		req.URL.Path = suffixPath
		fs.ServeHTTP(w, req)
	})
}

func (r *Router) AddRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}
