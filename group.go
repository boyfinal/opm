package opm

import "net/http"

type Group struct {
	prefix     string
	middleware []MiddlewareFunc
	router     *Router
}

func (g *Group) Prefix(prefix string) *Group {
	g.prefix = prefix
	return g
}

func (g *Group) Path(path string) *Route {
	return g.router.NewRoute().Path(path)
}

func (g *Group) GET(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodGet, path, h, m...)
}

func (g *Group) HEAD(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodHead, path, h, m...)
}

func (g *Group) POST(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodPost, path, h, m...)
}

func (g *Group) PUT(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodPut, path, h, m...)
}

func (g *Group) PATCH(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodPatch, path, h, m...)
}

func (g *Group) DELETE(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodDelete, path, h, m...)
}

func (g *Group) CONNECT(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodConnect, path, h, m...)
}

func (g *Group) OPTIONS(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodOptions, path, h, m...)
}

func (g *Group) TRACE(path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(http.MethodTrace, path, h, m...)
}

func (g *Group) Any(path string, h Handler, m ...MiddlewareFunc) []*Route {
	routers := make([]*Route, len(methods))

	for i, method := range methods {
		routers[i] = g.add(method, path, h, m...)
	}

	return routers
}

func (g *Group) Add(method, path string, h Handler, m ...MiddlewareFunc) *Route {
	return g.add(method, path, h, m...)
}

func (g *Group) add(method, path string, h Handler, middleware ...MiddlewareFunc) *Route {
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.router.Path(g.prefix + path).Handler(h).Method(method).Use(m...)
}

func (g *Group) File(path, file string) {
	h := HandlerFunc(func(c Context) error {
		return c.File(file)
	})

	g.Path(g.prefix + path).Handler(h).Method(http.MethodGet)
}

func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	g.router.Group(g.prefix+prefix, m...)
	return g.router.Group(g.prefix+prefix, m...)
}

func (r *Router) Prefix(prefix string) *Group {
	return r.Group(prefix)
}

func (r *Router) Middleware(m ...MiddlewareFunc) *Group {
	return r.Group("", m...)
}

func (g *Group) Routes(f func(*Group) *Group) *Group {
	f(g)

	return g
}
