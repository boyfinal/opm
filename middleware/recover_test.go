package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boyfinal/opm"
)

func TestRecover(t *testing.T) {
	or := opm.NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := or.NewContext(rec, req)
	h := Recover(func() opm.Handler {
		return opm.HandlerFunc(func(c opm.Context) error {
			panic("test")
		})
	}())

	h.Run(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("%v", rec.Code)
	}
}
