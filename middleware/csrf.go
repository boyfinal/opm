package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/boyfinal/opm"
	"github.com/gorilla/securecookie"
)

const tokenLength = 32

const (
	tokenKey     string = "csrf.Token"
	formKey      string = "csrf.Form"
	errorKey     string = "csrf.Error"
	skipCheckKey string = "csrf.Skip"
	cookieName   string = "__csrf"
)

var (
	fieldName   = tokenKey
	defaultAge  = 3600 * 12
	headerName  = "X-CSRF-Token"
	safeMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

	sc = securecookie.New([]byte("32-byte-long-auth-key"), nil)
)

type CSRFConfig struct {
	MaxAge   int
	Domain   string
	Path     string
	HTTPOnly bool
	Secure   bool
	SameSite http.SameSite

	RequestHeader string
	FieldName     string
	CookieName    string
}

func CSRF(config CSRFConfig) opm.MiddlewareFunc {
	if config.MaxAge == 0 {
		config.MaxAge = defaultAge
	}

	if config.RequestHeader == "" {
		config.RequestHeader = headerName
	}

	if config.FieldName == "" {
		config.FieldName = fieldName
	}

	if config.CookieName == "" {
		config.CookieName = cookieName
	}

	if config.SameSite == 0 {
		config.SameSite = http.SameSiteDefaultMode
	}

	if config.SameSite == http.SameSiteNoneMode {
		config.Secure = true
	}

	return func(next opm.Handler) opm.Handler {
		return opm.HandlerFunc(func(c opm.Context) error {
			realToken, err := getToken(config.CookieName, c.Request())
			if err != nil || len(realToken) != tokenLength {
				realToken, err = opm.GenerateRandBytes(tokenLength)
				if err != nil {
					return c.String(http.StatusBadRequest, err.Error())
				}

				encoded, err := sc.Encode(config.CookieName, realToken)
				if err != nil {
					return c.String(http.StatusBadRequest, err.Error())
				}

				cookie := &http.Cookie{
					Name:     config.CookieName,
					Value:    encoded,
					MaxAge:   config.MaxAge,
					HttpOnly: config.HTTPOnly,
					Secure:   config.Secure,
					SameSite: config.SameSite,
					Path:     config.Path,
					Domain:   config.Domain,
					Expires:  time.Now().Add(time.Duration(config.MaxAge) * time.Second),
				}

				c.SetCookie(cookie)
			}

			c.Set("csrf", mask(realToken))

			if !opm.InArrayString(safeMethods, c.Request().Method) {
				if realToken == nil {
					return c.String(http.StatusBadRequest, "invalid csrf token")
				}

				requestToken := unmask(requestToken(c.Request()))
				if !compareTokens(requestToken, realToken) {
					return c.String(http.StatusBadRequest, "invalid csrf token")
				}
			}

			c.Response().Header().Add("Vary", "Cookie")
			return next.Run(c)
		})
	}
}

func requestToken(r *http.Request) []byte {
	issued := r.Header.Get(headerName)

	if issued == "" {
		issued = r.PostFormValue(fieldName)
	}

	if issued == "" && r.MultipartForm != nil {
		vals := r.MultipartForm.Value[fieldName]

		if len(vals) > 0 {
			issued = vals[0]
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(issued)
	if err != nil {
		return nil
	}

	return decoded
}

func mask(realToken []byte) string {
	otp, err := opm.GenerateRandBytes(tokenLength)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(append(otp, opm.Xor(otp, realToken)...))
}

func unmask(issued []byte) []byte {
	if len(issued) != tokenLength*2 {
		return nil
	}

	otp := issued[tokenLength:]
	masked := issued[:tokenLength]

	return opm.Xor(otp, masked)
}

func compareTokens(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(a, b) == 1
}

func getToken(name string, r *http.Request) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}

	token := make([]byte, tokenLength)
	err = sc.Decode(name, cookie.Value, &token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
