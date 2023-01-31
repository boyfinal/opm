package middleware

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/boyfinal/opm"
)

// Logger --
func Logger(next opm.Handler) opm.Handler {
	return opm.Handler(func(c opm.Context) error {
		defer func() {
			startedTime := time.Now()
			randomNumber := rand.Intn(100000-1000) + 1000
			requestID := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d + %d", randomNumber, startedTime.Nanosecond()))))

			c.Response().Header().Set(opm.HeaderXRequestID, requestID)
			log.Println(requestID, c.Request().Method, c.Request().URL.Path, c.RealIP(), c.Request().UserAgent(), c.Route().GetName())
		}()

		return next(c)
	})
}
