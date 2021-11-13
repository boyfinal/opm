package middleware

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"opm"
	"time"
)

// Logger --
func Logger(next opm.Handler) opm.Handler {
	return opm.HandlerFunc(func(c opm.Context) error {
		defer func() {
			startedTime := time.Now()
			randomNumber := rand.Intn(100000-1000) + 1000
			requestID := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d + %d", randomNumber, startedTime.Nanosecond()))))

			c.Response().Header().Set("X-Request-ID", requestID)
			log.Println(requestID, c.Request().Method, c.Request().URL.Path, c.Request().RemoteAddr, c.Request().UserAgent(), c.Route().GetName())
		}()

		return next.Run(c)
	})
}
