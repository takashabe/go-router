package router

type Routing struct{}

func (r *Routing) Lookup(path string) RouteHandler {}

func (r *Routing) Construct([]*Route) {}
