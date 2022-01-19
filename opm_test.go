package opm

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEcho(t *testing.T) {
	r := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := r.NewContext(rec, req)

	// Router
	r.NewRoute()

	// DefaultHTTPErrorHandler
	r.DefaultHTTPErrorHandler(errors.New("error"), c)
	if http.StatusInternalServerError != rec.Code {
		t.Error(rec.Code)
	}
}

func TestEchoNotFound(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/tests", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Error(rec.Code)
	}
}

func TestEchoNotAllow(t *testing.T) {
	r := NewRouter()

	r.GET("/test", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Error(rec.Code)
	}
}

func TestGet(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodGet, "/test", o)
}

func TestHead(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodHead, "/test", o)
}

func TestPost(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodPost, "/test", o)
}

func TestPut(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodPut, "/test", o)
}

func TestPatch(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodPatch, "/test", o)
}

func TestDelete(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodDelete, "/test", o)
}

func TestConnect(t *testing.T) {
	o := NewRouter()
	testMethod(t, http.MethodConnect, "/test", o)
}

func testMethod(t *testing.T, method, path string, o *Router) {
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, method)
	}))

	i := interface{}(o)
	reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{p, h})

	code, body := request(method, path, o)
	if method != body {
		t.Error(code, body, method)
	}
}

func request(method, path string, o *Router) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()

	o.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
