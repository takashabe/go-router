package router

type Route struct {
	// called on request(ServeHTTP). behavior like http.handler
	method  string
	path    string
	handler baseHandler
}

func (r *Route) HandleFunc(method, path string, h baseHandler) *Route {
	r.method = method
	r.path = path
	r.handler = h
	return r
}
