package router

type RouteHandler interface{}

type Route struct {
	handler RouteHandler
	path    string
}

func (r *Route) Path(path string) *Route {
	r.path = path
	return r
}

func (r *Route) HandleFunc(path string, h RouteHandler) *Route {
	r.handler = h
	return r
}
