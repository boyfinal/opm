package opm

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type (
	Context interface {
		// Request returns `*http.Request`
		Request() *http.Request

		// SetRequest sets `*http.Request`
		SetRequest(*http.Request)

		// Response returns `http.ResponseWriter`
		Response() http.ResponseWriter

		// SetResponse sets `http.ResponseWriter`
		SetResponse(http.ResponseWriter)

		// Logger returns the Looger instance
		Logger() Logger

		// SetLogger set the logger
		SetLogger(l Logger)

		// SetValue saves data in body response
		SetValue(string, interface{})

		// GetValue data in body response
		GetValue(string) interface{}

		// GetValues data in body response
		GetValues() map[string]interface{}

		// Param returns path parameter by name.
		Param(name string) string

		// ParamNames sets path parameter names.
		SetParamNames(names ...string)

		// ParamNames returns path parameter names.
		ParamNames() []string

		// ParamNames sets path parameter values.
		SetParamValues(values ...string)

		// ParamNames returns path parameter values.
		ParamValues() []string

		// QueryParam returns the query param for the provided.
		QueryParam(name string) string

		// QueryParams returns the query parameters as `url.Values`.
		QueryParams() url.Values

		// QueryString returns the URL query string.
		QueryString() string

		// FormValue returns the form field value for the provided name.
		FormValue(name string) string

		// FormFile returns the multipart form file for the provided name.
		FormFile(name string) (*multipart.FileHeader, error)

		// HTML sends a blod response with  content type and status code
		Blob(code int, contentType string, b []byte) (err error)

		// Decode reads the next JSON-encoded value from request
		Decode(interface{}) error

		// Stream sends a streaming response with status code and content type.
		Stream(code int, contentType string, r io.Reader) error

		// HTML sends a HTTP response with status code
		HTML(code int, html string) (err error)

		// HTMLBlob sends a HTTP blod response with status code
		HTMLBlob(code int, b []byte) (err error)

		// Render renders a template with data and sends a text/html response with
		// status code.
		Render(code int, filename string) error

		// JSON sends a JSON response with status code
		JSON(code int, data interface{}) error

		// JSON sends a JSON response with status code
		String(code int, data string) error

		// Redirect Redirects the request to provider URL with status code
		Redirect(code int, url string) error

		// File sends a file response
		File(file string) error

		// NoContent sends a response with no body anh a status code
		NoContent(code int) error

		// Get retrieves data from the body response.
		Get(key string) interface{}

		// Set saves data in the body response.
		Set(key string, val interface{})

		// Get data in the body response.
		Body() map[string]interface{}

		Reset(w http.ResponseWriter, r *http.Request)

		// Cookie returns the named cookie provided in the request.
		Cookie(name string) (*http.Cookie, error)

		// SetCookie adds a `Set-Cookie` header in HTTP response.
		SetCookie(cookie *http.Cookie)

		// Cookies returns the HTTP cookies sent with the request.
		Cookies() []*http.Cookie

		// Renderer sets service provide response HTML.
		SetRenderer(Renderer)

		// Renderer returns service provide response HTML.
		Renderer() Renderer

		// SetRoute sets `*Route`
		SetRoute(*Route)

		// Route returns `*Route`
		Route() *Route

		// Domain returns site domain
		Domain() string

		// RealIP return real ip
		RealIP() string
	}

	Renderer interface {
		Render(io.Writer, string, interface{}) error
	}

	context struct {
		lock     sync.RWMutex
		request  *http.Request
		response http.ResponseWriter
		logger   Logger
		query    url.Values
		renderer Renderer
		pnames   []string
		pvalues  []string
		body     map[string]interface{}
		route    *Route
	}
)

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *context) Response() http.ResponseWriter {
	return c.response
}

func (c *context) SetResponse(w http.ResponseWriter) {
	c.response = w
}

func (c *context) SetLogger(l Logger) {
	c.logger = l
}

func (c *context) Logger() Logger {
	return c.logger
}

func (c *context) SetValue(key string, val interface{}) {
	c.body[key] = val
}

func (c *context) GetValue(key string) interface{} {
	return c.body[key]
}

func (c *context) GetValues() map[string]interface{} {
	return c.body
}

func (c *context) Param(key string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == key {
				return c.pvalues[i]
			}
		}
	}

	return ""
}

func (c *context) ParamNames() []string {
	return c.pnames
}

func (c *context) SetParamNames(names ...string) {
	c.pnames = names
}

func (c *context) SetParamValues(values ...string) {
	c.pvalues = values[:len(c.pnames)]
}

func (c *context) ParamValues() []string {
	return c.pvalues
}

func (c *context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.Request().URL.Query()
	}

	return c.query.Get(name)
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.Request().URL.Query()
	}

	return c.query
}

func (c *context) QueryString() string {
	return c.Request().URL.RawQuery
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	f, fh, err := c.Request().FormFile(name)
	if err != nil {
		return nil, err
	}

	f.Close()
	return fh, nil
}

func (c *context) FormValue(name string) string {
	return c.Request().FormValue(name)
}

func (c *context) FormParams(name string) (url.Values, error) {
	return c.Request().Form, nil
}

func (c *context) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *context) Render(code int, filename string) (err error) {
	buf := new(bytes.Buffer)
	if err = c.Renderer().Render(buf, filename, c.Body()); err != nil {
		return
	}

	return c.HTMLBlob(code, buf.Bytes())
}

func (c *context) HTML(code int, html string) (err error) {
	return c.HTMLBlob(code, []byte(html))
}

func (c *context) HTMLBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (c *context) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.Response().WriteHeader(code)
	_, err = c.Response().Write(b)
	return
}

func (c *context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.Response().WriteHeader(code)
	_, err = io.Copy(c.response, r)
	return
}

func (c *context) String(code int, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *context) json(code int, b []byte) error {
	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(code)
	_, err := c.Response().Write(b)
	return err
}

func (c *context) JSON(code int, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.json(code, b)
}

func (c *context) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response = w
	c.query = nil
	c.pnames = nil
	c.pvalues = nil
	c.route = nil
	c.body = make(map[string]interface{})
}

func (c *context) NoContent(code int) error {
	c.Response().WriteHeader(code)
	return nil
}

func (c *context) Redirect(code int, url string) error {
	http.Redirect(c.Response(), c.Request(), url, code)
	return nil
}

// Get retrieves data from the body response.
func (c *context) Get(key string) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.body == nil {
		return nil
	}

	return c.body[key]
}

// Set saves data in the body response.
func (c *context) Set(key string, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.body == nil {
		c.body = make(map[string]interface{})
	}

	c.body[key] = val
}

// Get data in the body response.
func (c *context) Body() map[string]interface{} {
	return c.body
}

func (c *context) File(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}

	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return ErrNotFound
	}

	if d.IsDir() {
		file = filepath.Join(file, "index.html")
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		defer f.Close()

		if d, err = f.Stat(); err != nil {
			return ErrNotFound
		}
	}

	http.ServeContent(c.Response(), c.Request(), d.Name(), d.ModTime(), f)
	return nil
}

// Cookie returns the named cookie provided in the request.
func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *context) Decode(v interface{}) error {
	return json.NewDecoder(c.Request().Body).Decode(v)
}

func (c *context) SetRenderer(v Renderer) {
	c.renderer = v
}

func (c *context) Renderer() Renderer {
	return c.renderer
}

func (c *context) Route() *Route {
	return c.route
}

func (c *context) SetRoute(route *Route) {
	c.route = route
}

func (c *context) Domain() string {
	return c.Request().Host
}

func (c *context) RealIP() string {
	ip := c.Request().Header.Get(HeaderXRealIP)
	if netIP := net.ParseIP(ip); netIP != nil {
		return ip
	}

	ips := c.Request().Header.Get(HeaderXForwardedFor)
	splitIP := strings.Split(ips, ",")
	for _, ip := range splitIP {
		if netIP := net.ParseIP(ip); netIP != nil {
			return ip
		}
	}

	rip, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
	return rip
}
