package middleware

import (
	"log"
	"net/http"

	"github.com/boyfinal/opm"
)

type Limiter struct {
	MaxLimit int
	IPCount  map[string]int
}

func ProtectLimiter(max int) *Limiter {
	return &Limiter{MaxLimit: max, IPCount: make(map[string]int)}
}

func (m *Limiter) Middleware(next opm.Handler) opm.Handler {
	return opm.Handler(func(c opm.Context) error {
		if m.MaxLimit == 0 {
			return next(c)
		}

		// Get the IP address for the current user.
		ip := c.RealIP()
		if ip == "" {
			return next(c)
		}

		m.IPCount[ip]++
		log.Println(m.IPCount[ip])

		defer func() {
			m.IPCount[ip]--
		}()

		if m.IPCount[ip] > m.MaxLimit {
			return c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		}

		return next(c)
	})
}
