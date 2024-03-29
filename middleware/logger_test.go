package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boyfinal/opm"
)

func TestLogger(t *testing.T) {
	or := opm.Make()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := or.NewContext(rec, req)
	h := Logger(func() opm.Handler {
		return opm.Handler(func(c opm.Context) error {
			return c.String(http.StatusOK, "test")
		})
	}())

	// Status 2xx
	h(c)

	// Status 3xx
	rec = httptest.NewRecorder()
	c = or.NewContext(rec, req)
	h = Logger(func() opm.Handler {
		return opm.Handler(func(c opm.Context) error {
			return c.String(http.StatusTemporaryRedirect, "test")
		})
	}())

	h(c)

	// Status 4xx
	rec = httptest.NewRecorder()
	c = or.NewContext(rec, req)
	h = Logger(func() opm.Handler {
		return opm.Handler(func(c opm.Context) error {
			return c.String(http.StatusNotFound, "test")
		})
	}())

	h(c)

	// Status 5xx with empty path
	rec = httptest.NewRecorder()
	c = or.NewContext(rec, req)
	h = Logger(func() opm.Handler {
		return opm.Handler(func(c opm.Context) error {
			return errors.New("error")
		})
	}())

	h(c)
}
