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
// It returns a new session if the session doesn`t exist.
func (cs *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	var err error
	var ctx = r.Context()
	sesValue := ctx.Value(sessionKeyInContext)
	session, ok := sesValue.(Session) // type assertion
	if ok {
		return &session, nil
	}

	newSession := NewSession(cs, name)
	newSession.IsNew = true
	cookie, err := r.Cookie(name)
	if err == nil {
		err = cs.Codec.Decode(name, cookie.Value, &newSession.Values)
		if err == nil {
			newSession.IsNew = false
		}
	}

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
