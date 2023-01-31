package opm

import "fmt"

type (
	Route struct {
		path        string
		name        string
		method      string
		err         error
		handler     Handler
		reg         *routeRegexp
		middleware  []MiddlewareFunc
		namedRoutes map[string]*Route
	}

	RouteList  []*Route
	RouteNames map[string]*Route
)

// Method sets type method `*Route`
func (r *Route) Method(method string) *Route {
	r.method = method
	return r
}

// Method gets name `*Route`
func (r *Route) GetName() string {
	if r == nil {
		return ""
	}

	return r.name
}

func (r *Route) Name(name string) *Route {
	if r.namedRoutes[name] != nil {
		r.err = fmt.Errorf("route already has name %s", name)
	}

	if r.err == nil {
		r.name = name
		r.namedRoutes[name] = r
	}
	return r
}

func (r *Route) GetPath() string {
	return r.path
}

func (r *Route) Path(path string) *Route {
	rr, err := newRouteRegexp(path)
	r.reg = rr
	r.err = err
	r.path = path
	if r.name == "" {
		r.name = path
	}

	return r
}

func (r *Route) Handler(handler Handler) *Route {
	if r.err == nil {
		r.handler = handler
	}

	return r
}

func (r *Route) HandlerFunc(f func(Context) error) *Route {
	return r.Handler(f)
}

func (r *Route) Use(m ...MiddlewareFunc) *Route {
	r.middleware = append(r.middleware, m...)
	return r
}

func extractVars(input string, matches []int, names []string) []string {
	var values []string
	for i := range names {
		values = append(values, input[matches[2*i+2]:matches[2*i+3]])
	}

	return values
}
