package opm

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrTooManyRequests             = NewHTTPError(http.StatusTooManyRequests)
	ErrBadRequest                  = NewHTTPError(http.StatusBadRequest)
	ErrBadGateway                  = NewHTTPError(http.StatusBadGateway)
	ErrInternalServerError         = NewHTTPError(http.StatusInternalServerError)
	ErrRequestTimeout              = NewHTTPError(http.StatusRequestTimeout)
	ErrServiceUnavailable          = NewHTTPError(http.StatusServiceUnavailable)
	ErrValidatorNotRegistered      = errors.New("validator not registered")
	ErrRendererNotRegistered       = errors.New("renderer not registered")
	ErrInvalidRedirectCode         = errors.New("invalid redirect status code")
	ErrCookieNotFound              = errors.New("cookie not found")
	ErrInvalidCertOrKeyType        = errors.New("invalid cert or key type, must be string or []byte")
	ErrInvalidListenerNetwork      = errors.New("invalid listener network")

	methods = [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
)

type (
	Router struct {
		sync.Mutex

		Logger      Logger
		Renderer    Renderer
		pool        sync.Pool
		namedRoutes map[string]*Route
		routes      []*Route
		middleware  []MiddlewareFunc

		NotFoundHandler         Handler
		SystemErrorHandler      Handler
		MethodNotAllowedHandler Handler
	}

	RouteMatch struct {
		Route    *Route
		Handler  Handler
		PNames   []string
		PValues  []string
		MatchErr error
	}

	Handler interface {
		Run(Context) error
	}

	HandlerFunc    func(Context) error
	MiddlewareFunc func(Handler) Handler

	HTTPError struct {
		Code    int         `json:"-"`
		Message interface{} `json:"message"`
	}

	Map map[string]interface{}
)

func (f HandlerFunc) Run(c Context) error {
	return f(c)
}

var (
	notFoundHandler = HandlerFunc(func(c Context) error {
		return c.String(http.StatusNotFound, "404 page not found")
	})

	methodNotAllowedHandler = HandlerFunc(func(c Context) error {
		return c.NoContent(http.StatusMethodNotAllowed)
	})

	serverErrorHandler = HandlerFunc(func(c Context) error {
		return c.NoContent(http.StatusInternalServerError)
	})
)

// New returns a Server.
func NewRouter() *Router {
	r := &Router{
		namedRoutes: make(map[string]*Route),
	}

	r.pool.New = func() interface{} {
		return r.NewContext(nil, nil)
	}

	return r
}

func (r *Router) SetLog(logger Logger) *Router {
	r.Logger = logger
	return r
}

// NewContext returns a Context instance.
func (r *Router) NewContext(w http.ResponseWriter, req *http.Request) Context {
	return &context{
		request:  req,
		response: w,
		renderer: r.Renderer,
		logger:   r.Logger,
		body:     make(map[string]interface{}),
	}
}

func (r *Router) Use(m ...MiddlewareFunc) {
	r.middleware = append(r.middleware, m...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	c := r.pool.Get().(*context)
	c.Reset(w, rq)

	var h Handler
	var match RouteMatch
	if r.Match(rq, &match) {
		h = match.Handler
		h = applyMiddleware(h, r.middleware...)

		c.SetRoute(match.Route)
		c.SetParamNames(match.PNames...)
		c.SetParamValues(match.PValues...)
	}

	if h == nil && match.MatchErr == ErrMethodNotAllowed {
		h = methodNotAllowedHandler
	}

	if h == nil {
		h = notFoundHandler
	}

	if err := h.Run(c); err != nil {
		if err == ErrNotFound {
			if r.NotFoundHandler != nil {
				r.NotFoundHandler.Run(c)
			} else {
				notFoundHandler.Run(c)
			}
		} else {
			if r.Logger != nil {
				r.Logger.Error(err)
			} else {
				fmt.Println(err)
			}

			if r.SystemErrorHandler != nil {
				r.SystemErrorHandler.Run(c)
			} else {
				serverErrorHandler.Run(c)
			}
		}
	}

	r.pool.Put(c)
}

// NewRoute create new a Route
func (r *Router) NewRoute() *Route {
	route := &Route{namedRoutes: r.namedRoutes, middlewares: make([]MiddlewareFunc, 0)}
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Name(name string) *Route {
	return r.NewRoute().Name(name)
}

func (r *Router) Path(path string) *Route {
	return r.NewRoute().Path(path)
}

func (r *Router) Match(req *http.Request, match *RouteMatch) bool {
	for _, route := range r.routes {
		if route.err != nil || route.reg == nil {
			continue
		}

		if matched := route.reg.Math(req); !matched {
			continue
		}

		if route.method != req.Method {
			match.MatchErr = ErrMethodNotAllowed
			continue
		}

		match.Route = route
		match.Handler = route.handler
		match.MatchErr = route.err
		if len(route.middlewares) > 0 {
			match.Handler = applyMiddleware(match.Handler, route.middlewares...)
		}

		match.PNames = route.reg.VarsN

		path := getPath(req)
		matches := route.reg.regexp.FindStringSubmatchIndex(path)
		if len(matches) > 0 {
			match.PValues = extractVars(path, matches, route.reg.VarsN)
		}

		return true
	}

	if match.MatchErr == ErrMethodNotAllowed {
		if r.MethodNotAllowedHandler != nil {
			match.Handler = r.MethodNotAllowedHandler
			return true
		}

		return false
	}

	if r.NotFoundHandler != nil {
		match.Handler = r.NotFoundHandler
		match.MatchErr = ErrNotFound
		return true
	}

	match.MatchErr = ErrNotFound
	return false
}

func (r *Router) GET(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodGet, path, h, m...)
}

func (r *Router) HEAD(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodHead, path, h, m...)
}

func (r *Router) POST(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodPost, path, h, m...)
}

func (r *Router) PUT(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodPut, path, h, m...)
}

func (r *Router) PATCH(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodPatch, path, h, m...)
}

func (r *Router) DELETE(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodDelete, path, h, m...)
}

func (r *Router) CONNECT(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodConnect, path, h, m...)
}

func (r *Router) OPTIONS(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodOptions, path, h, m...)
}

func (r *Router) TRACE(path string, h Handler, m ...MiddlewareFunc) *Route {
	return r.add(http.MethodTrace, path, h, m...)
}

func (r *Router) Any(path string, h Handler, m ...MiddlewareFunc) []*Route {
	routers := make([]*Route, len(methods))

	for i, method := range methods {
		routers[i] = r.add(method, path, h, m...)
	}

	return routers
}

func (r *Router) add(method, path string, h Handler, m ...MiddlewareFunc) *Route {
	h = applyMiddleware(h, m...)
	return r.Path(path).Handler(h).Method(method)
}

func (r *Router) Static(path, root string) *Route {
	if root == "" {
		root = "."
	}

	return r.static(path, root, r.GET)
}

func (r *Router) static(path, root string, get func(string, Handler, ...MiddlewareFunc) *Route) *Route {
	f := HandlerFunc(func(c Context) error {
		p, err := url.PathUnescape(c.Param("path"))
		if err != nil {
			return err
		}

		name := filepath.Join(root, filepath.Clean("/"+p))
		fi, err := os.Stat(name)
		if err != nil {
			return ErrNotFound
		}

		p = c.Request().URL.Path
		if fi.IsDir() && p[len(p)-1] != '/' {
			return c.Redirect(http.StatusMovedPermanently, p+"/")
		}

		return c.File(name)
	})

	if !strings.HasSuffix(path, "/") {
		path = Strcon(path, "/")
	}

	path = Strcon(path, "{path:.*}")
	return get(path, f)
}

func (r *Router) File(path, file string) {
	h := HandlerFunc(func(c Context) error {
		return c.File(file)
	})

	r.Path(path).Handler(h).Method(http.MethodGet)
}

func applyMiddleware(h Handler, middleware ...MiddlewareFunc) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func getPath(r *http.Request) string {
	path := r.URL.EscapedPath()
	if path == "" {
		path = r.URL.Path
	}

	return path
}

func (r *Router) DefaultHTTPErrorHandler(err error, c Context) {
	he, ok := err.(*HTTPError)
	if !ok {
		he = &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	// Issue #1426
	code := he.Code
	message := he.Message
	if m, ok := he.Message.(string); ok {
		message = Map{"message": m}
	}

	if c.Request().Method == http.MethodHead {
		err = c.NoContent(he.Code)
	} else {
		err = c.JSON(code, message)
	}

	if err != nil {
		r.Logger.Error(err)
	}
}

func (r *Router) Group(prefix string, m ...MiddlewareFunc) *Group {
	return &Group{router: r, prefix: prefix, middlewares: m}
}

func NewHTTPError(code int, message ...interface{}) *HTTPError {
	err := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		err.Message = message[0]
	}

	return err
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
}
