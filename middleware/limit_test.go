package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boyfinal/opm"
)

func TestLimiter(t *testing.T) {
	or := opm.Make()

	mw := ProtectLimiter(1)
	handler := func() opm.Handler {
		return opm.HandlerFunc(func(c opm.Context) error {
			time.Sleep(1 * time.Second)
			return c.String(http.StatusOK, "test")
		})
	}

	testCases := []struct {
		id   string
		code int
	}{
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusTooManyRequests},
	}

	for _, tc := range testCases {
		go func(tc struct {
			id   string
			code int
		}) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add(opm.HeaderXRealIP, tc.id)
			rec := httptest.NewRecorder()
			c := or.NewContext(rec, req)
			mw.Middleware(handler()).Run(c)
			if tc.code != rec.Code {
				t.Errorf("%v - %v", tc.code, rec.Code)
			}

			t.Logf("%v - %v", tc.code, rec.Code)
		}(tc)
	}

	time.Sleep(1 * time.Second)
}
