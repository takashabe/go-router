package router

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// TokenParam represents parameter in the URL path
	// e.g. "/user/:id"
	TokenParam = ":"

	// TokenWildcard represents wildcard in the URL path
	// e.g. "/static/*filepath
	TokenWildcard = "*"
)

// Errors
var (
	ErrNotFoundHandler = errors.New("not found matched handler")
	ErrInvalidHandler  = errors.New("invalid handler")
	ErrInvalidParam    = errors.New("invalid param")
)

// Routing is represents routing tree
type Routing interface {
	Lookup(method, path string) (HandlerData, error)
	Insert(method, path string, handler baseHandler) error
}

// HandlerData is represents handler function and args
type HandlerData struct {
	handler baseHandler
	params  []interface{}
}

// ValidationParam is customize validation parameter for the baseHandler
type ValidationParam interface {
	Validate(raw string) bool
}

type baseHandler interface{}

// Router is represents routing algorism and routes
type Router struct {
	NotFoundHandler http.Handler
	Routing         Routing
	routes          []*Route
	outLog          *log.Logger
	errLog          *log.Logger
}

// NewRouter return created Router
func NewRouter() *Router {
	return &Router{
		NotFoundHandler: http.NotFoundHandler(),
		Routing:         NewTrie(),
		outLog:          newLogger(os.Stdout),
		errLog:          newLogger(os.Stderr),
	}
}

func newLogger(w io.Writer) *log.Logger {
	return log.New(w, "", log.LstdFlags|log.Lshortfile)
}

// SetOutLogger setting normal logger
func (r *Router) SetOutLogger(w io.Writer) {
	r.outLog = newLogger(w)
}

// SetErrLogger setting error logger
func (r *Router) SetErrLogger(w io.Writer) {
	r.errLog = newLogger(w)
}

func (r *Router) accessLogf(format string, args ...interface{}) {
	if env := os.Getenv("GO_ROUTER_ENABLE_LOGGING"); len(env) != 0 {
		r.outLog.Printf("%s\n", args)
	}
}

func (r *Router) errorLogf(format string, args ...interface{}) {
	if env := os.Getenv("GO_ROUTER_ENABLE_LOGGING"); len(env) != 0 {
		r.errLog.Printf("[error] %s\n", args)
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
		param := hd.params[i-numStaticArgs]
		t := ref.In(i)
		switch t.Kind() {
		case reflect.Int:
			p, err := strconv.Atoi(param.(string))
			if err != nil {
				return nil, errors.Wrapf(ErrInvalidParam, "path=%s, error=%s", req.URL.Path, err)
			}
			args = append(args, reflect.ValueOf(p))
		case reflect.Ptr:
			v, ok := reflect.New(t.Elem()).Interface().(ValidationParam)
			if !ok {
				return nil, errors.Wrapf(ErrInvalidParam, "Validatable parameters only allow ValidationParam, path=%s", req.URL.Path)
			}
			if !v.Validate(hd.params[i-numStaticArgs].(string)) {
				return nil, errors.Wrapf(ErrInvalidParam, "Failed to validation, path=%s", req.URL.Path)
			}
			args = append(args, reflect.ValueOf(v))
		default:
			args = append(args, reflect.ValueOf(param))
		}
	}
	return args, nil
}

// Get register handler via GET
func (r *Router) Get(path string, h baseHandler) *Route { return r.HandleFunc("GET", path, h) }

// Head register handler via HEAD
func (r *Router) Head(path string, h baseHandler) *Route { return r.HandleFunc("HEAD", path, h) }

// Post register handler via POST
func (r *Router) Post(path string, h baseHandler) *Route { return r.HandleFunc("POST", path, h) }

// Put register handler via PUT
func (r *Router) Put(path string, h baseHandler) *Route { return r.HandleFunc("PUT", path, h) }

// Patch register handler via PATCH
func (r *Router) Patch(path string, h baseHandler) *Route { return r.HandleFunc("PATCH", path, h) }

// Delete register handler via DELETE
func (r *Router) Delete(path string, h baseHandler) *Route { return r.HandleFunc("DELETE", path, h) }

// Options register handler via OPTIONS
func (r *Router) Options(path string, h baseHandler) *Route { return r.HandleFunc("OPTIONS", path, h) }

// HandleFunc register handler each HTTP method
func (r *Router) HandleFunc(method, path string, h baseHandler) *Route {
	route := r.AddRoute().HandleFunc(method, path, h)
	err := r.Routing.Insert(route.method, route.path, route.handler)
	if err != nil {
		r.errorLogf("failed registered path. path=%s, error=%v", path, err)
	}
	return route
}

// ServeDir register handler for static directories
func (r *Router) ServeDir(path string, root http.FileSystem) {
	fs := http.FileServer(root)
	r.Get(path, func(w http.ResponseWriter, req *http.Request, suffixPath string) {
		req.URL.Path = suffixPath
		fs.ServeHTTP(w, req)
	})
}

// ServeFile register handler for static files
func (r *Router) ServeFile(path string, file string) {
	r.Get(path, func(w http.ResponseWriter, req *http.Request) {
		if containsDotDot(file) {
			http.Error(w, "invalid URL path", http.StatusBadRequest)
			return
		}
		http.ServeFile(w, req, file)
	})
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

// AddRoute add route in router
func (r *Router) AddRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}

// PrintRoutes display all registered routes
func (r *Router) PrintRoutes(w io.Writer) {
	routes := r.routes
	list := new(bytes.Buffer)
	for _, route := range routes {
		ref := reflect.ValueOf(route.handler)
		method := runtime.FuncForPC(ref.Pointer()).Name()
		list.WriteString(fmt.Sprintf("[%s] \"%s\" -> %s\n",
			route.method, route.path, method))
	}
	fmt.Fprintln(w, list)
}
