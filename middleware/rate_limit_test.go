package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boyfinal/opm"
)

func TestRateLimiter(t *testing.T) {
	or := opm.Make()

	mw := RateLimiter(1, 3)
	handler := func() opm.Handler {
		return opm.Handler(func(c opm.Context) error {
			return c.String(http.StatusOK, "test")
		})
	}

	testCases := []struct {
		id   string
		code int
	}{
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusTooManyRequests},
		{"127.0.0.1", http.StatusTooManyRequests},
		{"127.0.0.1", http.StatusTooManyRequests},
		{"127.0.0.1", http.StatusTooManyRequests},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(opm.HeaderXRealIP, tc.id)

		rec := httptest.NewRecorder()
		c := or.NewContext(rec, req)

		h := mw.Middleware(handler())
		h(c)

		if tc.code != rec.Code {
			t.Errorf("%v - %v", tc.code, rec.Code)
		}
	}
}
