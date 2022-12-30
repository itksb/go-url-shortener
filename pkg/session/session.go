package session

import "net/http"

// Session stores the values and optional configuration for a session.
type Session struct {
	IsNew bool
	// Values contains the user-data for the session.
	Values  map[interface{}]interface{}
	Options *Options

	store Store
	name  string
}

// NewSession is a new session constructor
func NewSession(store Store, name string) *Session {
	return &Session{
		IsNew:   true,
		Values:  make(map[interface{}]interface{}),
		Options: NewOptions(),

		store: store,
		name:  name,
	}
}

// Name returns the name used to register the session.
func (s *Session) Name() string {
	return s.name
}

// Save - save this session  to the store (like CookieStore)
// Call Save before starting return http body
func (s *Session) Save(r *http.Request, w http.ResponseWriter) error {
	return s.store.Save(r, w, s)
}
