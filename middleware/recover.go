package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/boyfinal/opm"
)

// Recover --
func Recover(next opm.Handler) opm.Handler {
	return opm.Handler(func(c opm.Context) error {
		defer func() {
			var err error
			if r := recover(); r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}

				log.Println("Panic: ", err, "url", c.Request().URL, "raw_error", r)
				c.String(http.StatusInternalServerError, err.Error())
			}
		}()

		return next(c)
	})
}
