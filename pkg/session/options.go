package session

// Options stores configuration for a session or session store.
//
// Fields are a subset of http.Cookie fields.
// See details here: https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
type Options struct {
	// The Path attribute indicates a URL path that must exist
	// in the requested URL in order to send the Cookie header.
	Path   string
	Domain string
	// MaxAge=0 means no Max-Age attribute specified and the cookie will be
	// deleted after the browser session ends.
	// MaxAge<0 means delete cookie immediately.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge int
	// A cookie with the Secure attribute is only sent to the server
	// with an encrypted request over the HTTPS protocol.
	Secure bool
	// A cookie with the HTTPOnly attribute is inaccessible
	// to the JavaScript Document.cookie API; it's only sent to the server.
	HTTPOnly bool
}

// NewOptions - constructor
func NewOptions() *Options {
	return &Options{
		Path:     "/",
		MaxAge:   0,
		Secure:   false,
		HTTPOnly: true,
	}
}
