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
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

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
	Core struct {
		sync.Mutex

		Logger      Logger
		Renderer    Renderer
		pool        sync.Pool
		namedRoutes RouteNames
		routes      RouteList
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

	// Handler a responds to an HTTP request
	Handler interface {
		Run(Context) error
	}

	// HandlerFunc defines a function to serve HTTP request
	HandlerFunc func(Context) error

	// MiddlewareFunc defines a function to process middleware
	MiddlewareFunc func(Handler) Handler

	// HTTPError an error that occurred while handing a request
	HTTPError struct {
		Code    int         `json:"-"`
		Message interface{} `json:"message"`
	}

	Map map[string]interface{}
)

const (
	charsetUTF8 = "charset=UTF-8"

	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + ";" + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

func (f HandlerFunc) Run(c Context) error {
	return f(c)
}

var (
	// Default handler not found
	notFoundHandler = HandlerFunc(func(c Context) error {
		return c.String(http.StatusNotFound, "404 page not found")
	})

	// Default handler not allowed
	methodNotAllowedHandler = HandlerFunc(func(c Context) error {
		return c.NoContent(http.StatusMethodNotAllowed)
	})

	// Default handler error
	serverErrorHandler = HandlerFunc(func(c Context) error {
		return c.NoContent(http.StatusInternalServerError)
	})
)

// New returns a Server.
func Make() *Core {
	core := &Core{
		namedRoutes: make(map[string]*Route),
	}

	core.pool.New = func() interface{} {
		return core.NewContext(nil, nil)
	}

	return core
}

func (core *Core) SetLog(logger Logger) *Core {
	core.Logger = logger
	return core
}

// NewContext returns a Context instance.
func (core *Core) NewContext(w http.ResponseWriter, req *http.Request) Context {
	return &context{
		request:  req,
		response: w,
		renderer: core.Renderer,
		logger:   core.Logger,
		body:     make(map[string]interface{}),
	}
}

// Use adds middleware
func (core *Core) Use(m ...MiddlewareFunc) {
	core.middleware = append(core.middleware, m...)
}

func (core *Core) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	c := core.pool.Get().(*context)
	c.Reset(w, rq)

	var h Handler
	var match RouteMatch
	if core.Match(rq, &match) {
		h = match.Handler
		h = applyMiddleware(h, core.middleware...)

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
			if core.NotFoundHandler != nil {
				core.NotFoundHandler.Run(c)
			} else {
				notFoundHandler.Run(c)
			}
		} else {
			if core.Logger != nil {
				core.Logger.Error(err)
			} else {
				fmt.Println(err)
			}

			if core.SystemErrorHandler != nil {
				core.SystemErrorHandler.Run(c)
			} else {
				serverErrorHandler.Run(c)
			}
		}
	}

	core.pool.Put(c)
}

// NewRoute create new a Route
func (core *Core) NewRoute() *Route {
	route := &Route{namedRoutes: core.namedRoutes, middleware: make([]MiddlewareFunc, 0)}
	core.routes = append(core.routes, route)
	return route
}

func (core *Core) Name(name string) *Route {
	return core.NewRoute().Name(name)
}

func (core *Core) Path(path string) *Route {
	return core.NewRoute().Path(path)
}

func (core *Core) Match(req *http.Request, match *RouteMatch) bool {
	for _, route := range core.routes {
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

		if len(route.middleware) > 0 {
			match.Handler = applyMiddleware(match.Handler, route.middleware...)
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
		if core.MethodNotAllowedHandler != nil {
			match.Handler = core.MethodNotAllowedHandler
			return true
		}

		return false
	}

	if core.NotFoundHandler != nil {
		match.Handler = core.NotFoundHandler
		match.MatchErr = ErrNotFound
		return true
	}

	match.MatchErr = ErrNotFound
	return false
}

func (core *Core) GET(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodGet, path, h, m...)
}

func (core *Core) HEAD(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodHead, path, h, m...)
}

func (core *Core) POST(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodPost, path, h, m...)
}

func (core *Core) PUT(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodPut, path, h, m...)
}

func (core *Core) PATCH(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodPatch, path, h, m...)
}

func (core *Core) DELETE(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodDelete, path, h, m...)
}

func (core *Core) CONNECT(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodConnect, path, h, m...)
}

func (core *Core) OPTIONS(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodOptions, path, h, m...)
}

func (core *Core) TRACE(path string, h Handler, m ...MiddlewareFunc) *Route {
	return core.add(http.MethodTrace, path, h, m...)
}

func (core *Core) Any(path string, h Handler, m ...MiddlewareFunc) []*Route {
	routers := make([]*Route, len(methods))

	for i, method := range methods {
		routers[i] = core.add(method, path, h, m...)
	}

	return routers
}

func (core *Core) add(method, path string, h Handler, m ...MiddlewareFunc) *Route {
	h = applyMiddleware(h, m...)
	return core.Path(path).Handler(h).Method(method)
}

func (core *Core) Static(path, root string) *Route {
	if root == "" {
		root = "."
	}

	return core.static(path, root, core.GET)
}

func (core *Core) static(path, root string, get func(string, Handler, ...MiddlewareFunc) *Route) *Route {
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
		path = StrConcat(path, "/")
	}

	path = StrConcat(path, "{path:.*}")
	return get(path, f)
}

func (core *Core) File(path, file string) {
	h := HandlerFunc(func(c Context) error {
		return c.File(file)
	})

	core.Path(path).Handler(h).Method(http.MethodGet)
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

func (core *Core) DefaultHTTPErrorHandler(err error, c Context) {
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
		core.Logger.Error(err)
	}
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

func (core *Core) Run(addr string) {
	s := &Server{Addr: addr}
	s.Run()
}

func (core *Core) RunServer(s *Server) {
	s.Run()
}
