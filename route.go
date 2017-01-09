package router

type Route struct {
	// called on request(ServeHTTP). behavior like http.handler
	handler baseHandler
	path    string
}

func (r *Route) HandleFunc(path string, h baseHandler) *Route {
	r.path = path
	r.handler = h
	return r
}
