package opm

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOPM(t *testing.T) {
	r := Make()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := r.NewContext(rec, req)

	// Router
	r.NewRoute()

	// DefaultHTTPErrorHandler
	r.DefaultHTTPErrorHandler(errors.New("error"), c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code, fmt.Sprintf("status %d != %d", http.StatusInternalServerError, rec.Code))
}

func TestOPMNotFound(t *testing.T) {
	r := Make()

	req := httptest.NewRequest(http.MethodGet, "/tests", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, fmt.Sprintf("status %d != %d", http.StatusNotFound, rec.Code))
}

func TestOPMNotAllow(t *testing.T) {
	r := Make()

	r.GET("/test", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, fmt.Sprintf("status %d != %d", http.StatusMethodNotAllowed, rec.Code))
}

func TestGet(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodGet, "/test", o)
}

func TestHead(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodHead, "/test", o)
}

func TestPost(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodPost, "/test", o)
}

func TestPut(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodPut, "/test", o)
}

func TestPatch(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodPatch, "/test", o)
}

func TestDelete(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodDelete, "/test", o)
}

func TestConnect(t *testing.T) {
	o := Make()
	testMethod(t, http.MethodConnect, "/test", o)
}

func testMethod(t *testing.T, method, path string, core *Core) {
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, method)
	}))

	i := interface{}(core)
	reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{p, h})

	code, body := request(method, path, core)

	assert.Equal(t, body, method, fmt.Sprintf("code: %d, method: %s, body: %s", code, body, method))
}

func request(method, path string, core *Core) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()

	core.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
