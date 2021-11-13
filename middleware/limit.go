package middleware

import (
	"net"
	"net/http"
	"sync"

	"github.com/boyfinal/opm"
)

type Limiter struct {
	sync.Mutex
	MaxLimit int
	IPCount  map[string]int
}

func ProtectLimiter(max int) *Limiter {
	return &Limiter{MaxLimit: max}
}

func (m *Limiter) Middleware(next opm.Handler) opm.Handler {
	return opm.HandlerFunc(func(c opm.Context) error {
		if m.MaxLimit == 0 {
			return next.Run(c)
		}

		// Get the IP address for the current user.
		ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
		if err != nil {
			return err
		}

		// Get the # of times the visitor has visited in the last 60 seconds
		count, ok := m.IPCount[ip]
		if !ok {
			m.IPCount[ip] = 0
		}

		m.Lock()
		defer func() {
			println("end")
			m.Unlock()
			m.IPCount[ip]--
		}()

		if count > m.MaxLimit {
			return c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		}

		m.IPCount[ip]++
		return next.Run(c)
	})
}
