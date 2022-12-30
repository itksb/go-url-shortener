package session

import (
	"context"
	"net/http"
)

// Store interface
// compatible with Gorilla session (intentionally)
type Store interface {
	// Get should return a cached session.
	Get(r *http.Request, name string) (*Session, error)

	// Save should persist session to the underlying store implementation.
	Save(r *http.Request, w http.ResponseWriter, s *Session) error
}

// NewCookieStore returns a new CookieStore.
// Codec is a contract encapsulating encryption logic
func NewCookieStore(codec Codec) *CookieStore {
	cs := &CookieStore{
		Options: &Options{
			Path:     "/",
			Domain:   "",
			MaxAge:   60 * 60 * 24,
			Secure:   false,
			HTTPOnly: true,
		},
		Codec: codec,
	}
	return cs
}

// CookieStore stores sessions using secure cookies.
type CookieStore struct {
	Options *Options
	Codec   Codec
}

type sessionKey int

const sessionKeyInContext sessionKey = 0

// Get returns a session for the given name after adding it to the request context.
//
// It returns a session with empty values if the session values are not exist in the request context.
func (cs *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	var err error
	var ctx = r.Context()
	sesValue := ctx.Value(sessionKeyInContext)
	session, ok := sesValue.(Session) // type assertion
	if ok {                           // session exists in context? so just return it
		return &session, nil
	}
	// no session in the context, so create new one
	newSession := NewSession(cs, name)
	// maybe session saved in the cookie?
	cookie, err := r.Cookie(name)
	if err == nil { // cookie exists, so try to restore session values from the cookie
		err = cs.Codec.Decode(name, cookie.Value, &newSession.Values)
		if err == nil { // if no errors, then values restored correctly
			newSession.IsNew = false
		} else {
			// error
			return nil, err
		}
	} else {
		// named cookie not present, nothing values to restore
		newSession.IsNew = true
	}
	// write the created session to the context for further use.
	// So the next check on the context should return the same session.
	*r = *r.WithContext(context.WithValue(ctx, sessionKeyInContext, newSession))

	return newSession, nil
}

// Save adds s session to the response
func (cs *CookieStore) Save(r *http.Request, w http.ResponseWriter, s *Session) error {
	encoded, err := cs.Codec.Encode(s.Name(), s.Values)
	if err != nil {
		return err
	}

	options := cs.Options
	http.SetCookie(w, &http.Cookie{
		Name:     s.Name(),
		Value:    encoded,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
	})

	return nil
}
