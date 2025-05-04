package cstore

import (
	"net/http"
	"process/lib/config"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type SessionStore struct {
	Name  string
	Store *sessions.CookieStore
}

func NewStore(cfg *config.Config) *SessionStore {

	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	store := sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	store.Options = &sessions.Options{
		Path:     cfg.Session.Path,
		Domain:   cfg.Session.Domain,
		MaxAge:   3600 * cfg.Session.MaxAgeHours,
		HttpOnly: true,
		Secure:   cfg.InProduction,
		SameSite: http.SameSiteLaxMode,
	}

	return &SessionStore{
		Name:  cfg.Session.Name,
		Store: store,
	}
}
