package csrf

import (
	"net/url"

	"github.com/boyfinal/opm"
	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
)

// CSRF token length in bytes.
const tokenLength = 32

// Context/session keys & prefixes
const (
	tokenKey     string = "csrf.Token"
	formKey      string = "csrf.Form"
	skipCheckKey string = "csrf.Skip"
	cookieName   string = "__csrf"
)

var (
	// The name value used in form fields.
	fieldName = tokenKey
	// defaultAge sets the default MaxAge for cookies.
	defaultAge = 3600 * 12
	// The default HTTP request header to inspect
	headerName = "X-CSRF-Token"
	// Idempotent (safe) methods as defined by RFC7231 section 4.2.2.
	safeMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
)

// TemplateTag provides a default template tag - e.g. <input type="hidden" name="csrf.Token" value="{{.csrf}}"> - for use
// with the TemplateField function.
var TemplateTag = "csrfField"

var (
	// ErrNoReferer is returned when a HTTPS request provides an empty Referer
	// header.
	ErrNoReferer = errors.New("referer not supplied")
	// ErrBadReferer is returned when the scheme & host in the URL do not match
	// the supplied Referer header.
	ErrBadReferer = errors.New("referer invalid")
	// ErrNoToken is returned if no CSRF token is supplied in the request.
	ErrNoToken = errors.New("CSRF token not found in request")
	// ErrBadToken is returned if the CSRF token in the request does not match
	// the token in the session, or is otherwise malformed.
	ErrBadToken = errors.New("CSRF token invalid")
)

// SameSiteMode allows a server to define a cookie attribute making it impossible for
// the browser to send this cookie along with cross-site requests. The main
// goal is to mitigate the risk of cross-origin information leakage, and provide
// some protection against cross-site request forgery attacks.
//
// See https://tools.ietf.org/html/draft-ietf-httpbis-cookie-same-site-00 for details.
type SameSiteMode int

// SameSite options
const (
	// SameSiteDefaultMode sets the `SameSite` cookie attribute, which is
	// invalid in some older browsers due to changes in the SameSite spec. These
	// browsers will not send the cookie to the server.
	// csrf uses SameSiteLaxMode (SameSite=Lax) as the default as of v1.7.0+
	SameSiteDefaultMode SameSiteMode = iota + 1
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

type csrf struct {
	h    opm.Handler
	sc   *securecookie.SecureCookie
	st   store
	opts options
}

// options contains the optional settings for the CSRF middleware.
type options struct {
	MaxAge int
	Domain string
	Path   string
	// Note that the function and field names match the case of the associated
	// http.Cookie field instead of the "correct" HTTPOnly name that golint suggests.
	HttpOnly       bool
	Secure         bool
	SameSite       SameSiteMode
	RequestHeader  string
	FieldName      string
	CookieName     string
	TrustedOrigins []string
}

func Protect(authKey []byte, opts ...Option) func(opm.Handler) opm.Handler {
	return func(h opm.Handler) opm.Handler {
		cs := parseOptions(h, opts...)

		if cs.opts.MaxAge < 0 {
			// Default of 12 hours
			cs.opts.MaxAge = defaultAge
		}

		if cs.opts.FieldName == "" {
			cs.opts.FieldName = fieldName
		}

		if cs.opts.CookieName == "" {
			cs.opts.CookieName = cookieName
		}

		if cs.opts.RequestHeader == "" {
			cs.opts.RequestHeader = headerName
		}

		// Create an authenticated securecookie instance.
		if cs.sc == nil {
			cs.sc = securecookie.New(authKey, nil)
			// Use JSON serialization (faster than one-off gob encoding)
			cs.sc.SetSerializer(securecookie.JSONEncoder{})
			// Set the MaxAge of the underlying securecookie.
			cs.sc.MaxAge(cs.opts.MaxAge)
		}

		if cs.st == nil {
			// Default to the cookieStore
			cs.st = &cookieStore{
				name:     cs.opts.CookieName,
				maxAge:   cs.opts.MaxAge,
				secure:   cs.opts.Secure,
				httpOnly: cs.opts.HttpOnly,
				sameSite: cs.opts.SameSite,
				path:     cs.opts.Path,
				domain:   cs.opts.Domain,
				sc:       cs.sc,
			}
		}

		return cs.Handle
	}
}

// Implements http.Handler for the csrf type.
func (cs *csrf) Handle(c opm.Context) error {
	// Skip the check if directed to. This should always be a bool.
	if skip, ok := c.Get(skipCheckKey).(bool); ok {
		if skip {
			return cs.h(c)
		}
	}

	r := c.Request()
	w := c.Response()

	// Retrieve the token from the session.
	// An error represents either a cookie that failed HMAC validation
	// or that doesn't exist.
	realToken, err := cs.st.Get(r)
	if err != nil || len(realToken) != tokenLength {
		// If there was an error retrieving the token, the token doesn't exist
		// yet, or it's the wrong length, generate a new token.
		// Note that the new token will (correctly) fail validation downstream
		// as it will no longer match the request token.
		realToken, err = opm.GenerateRandBytes(tokenLength)
		if err != nil {
			return err
		}

		// Save the new (real) token in the session store.
		err = cs.st.Save(realToken, w)
		if err != nil {
			return err
		}
	}

	// Save the masked token to the request context
	c.Set(tokenKey, mask(realToken))
	c.Set(formKey, cs.opts.FieldName)

	// HTTP methods not defined as idempotent ("safe") under RFC7231 require
	// inspection.
	if !opm.InArrayString(safeMethods, r.Method) {
		// Enforce an origin check for HTTPS connections. As per the Django CSRF
		// implementation (https://goo.gl/vKA7GE) the Referer header is almost
		// always present for same-domain HTTP requests.
		if r.URL.Scheme == "https" {
			// Fetch the Referer value. Call the error handler if it's empty or
			// otherwise fails to parse.
			referer, err := url.Parse(r.Referer())
			if err != nil || referer.String() == "" {
				return err
			}

			valid := sameOrigin(r.URL, referer)

			if !valid {
				for _, trustedOrigin := range cs.opts.TrustedOrigins {
					if referer.Host == trustedOrigin {
						valid = true
						break
					}
				}
			}

			if !valid {
				return ErrBadReferer
			}
		}

		// If the token returned from the session store is nil for non-idempotent
		// ("unsafe") methods, call the error handler.
		if realToken == nil {
			return ErrNoToken
		}

		// Retrieve the combined token (pad + masked) token and unmask it.
		requestToken := unmask(cs.requestToken(r))

		// Compare the request token against the real token
		if !compareTokens(requestToken, realToken) {
			return ErrBadToken
		}

	}

	// Set the Vary: Cookie header to protect clients from caching the response.
	w.Header().Add("Vary", "Cookie")

	// Call the wrapped handler/router on success.
	return cs.h(c)
}
