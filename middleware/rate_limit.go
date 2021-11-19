package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/boyfinal/opm"
	"golang.org/x/time/rate"
)

type Visitor struct {
	*rate.Limiter
	Lastseen time.Time
}

type RateLimiterConfig struct {
	mu          sync.Mutex
	Rate        rate.Limit
	Burst       int
	visitors    map[string]*Visitor
	expiresIn   time.Duration
	lastCleanup time.Time
}

func RateLimiter(rate rate.Limit, b int) *RateLimiterConfig {
	return &RateLimiterConfig{
		Rate:        rate,
		Burst:       b,
		visitors:    make(map[string]*Visitor),
		lastCleanup: time.Now(),
		expiresIn:   1 * time.Second,
	}
}

func (m *RateLimiterConfig) Middleware(next opm.Handler) opm.Handler {
	return opm.HandlerFunc(func(c opm.Context) error {
		// Get the IP address for the current user.
		ip := c.RealIP()
		if ip == "" {
			return c.String(http.StatusNoContent, http.StatusText(http.StatusNoContent))
		}

		m.mu.Lock()

		limiter, exists := m.visitors[ip]
		if !exists {
			limiter = new(Visitor)
			limiter.Limiter = rate.NewLimiter(m.Rate, m.Burst)
			m.visitors[ip] = limiter
		}

		limiter.Lastseen = time.Now()
		m.refesh()

		m.mu.Unlock()

		if allow := limiter.AllowN(time.Now(), 1); !allow {
			return c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		}

		return next.Run(c)
	})
}

func (m *RateLimiterConfig) refesh() {
	if time.Since(m.lastCleanup) <= m.expiresIn {
		return
	}

	for id, visitor := range m.visitors {
		if time.Since(visitor.Lastseen) > m.expiresIn {
			delete(m.visitors, id)
		}
	}

	m.lastCleanup = time.Now()
}
