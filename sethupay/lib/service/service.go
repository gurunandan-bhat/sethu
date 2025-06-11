package service

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sethupay/lib/config"
	"sethupay/lib/model"

	mysqlstore "github.com/danielepintore/gorilla-sessions-mysql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Service struct {
	Config       *config.Config
	Model        *model.Model
	Muxer        *chi.Mux
	SessionStore *mysqlstore.MysqlStore
	Template     map[string]*template.Template
}

func NewService(cfg *config.Config) (*Service, error) {

	mux := chi.NewRouter()

	// force a redirect to https:// in production
	if cfg.InProduction {
		mux.Use(middleware.SetHeader(
			"Strict-Transport-Security",
			"max-age=63072000; includeSubDomains",
		))
	}

	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// csrfMiddleware := csrf.Protect(
	// 	[]byte(cfg.Security.CSRFKey),
	// 	csrf.Secure(cfg.InProduction),
	// 	csrf.SameSite(csrf.SameSiteStrictMode),
	// )
	// mux.Use(csrfMiddleware)

	model, err := model.NewModel(cfg)
	if err != nil {
		log.Fatalf("error initializing database connection: %s", err)
	}

	dbStore, err := model.NewDbSessionStore(cfg)
	if err != nil {
		log.Fatalf("error initializing db store: %s", err)
	}

	template, err := newTemplateCache(filepath.Join(cfg.AppRoot, "templates"))
	if err != nil {
		log.Fatalf("Cannot build template cache: %s", err)
	}

	s := &Service{
		Config:       cfg,
		SessionStore: dbStore,
		Model:        model,
		Muxer:        mux,
		Template:     template,
	}

	s.setRoutes()

	return s, nil
}

func (s *Service) setRoutes() {

	s.Muxer.Route("/sethupay", func(r chi.Router) {
		r.Method(http.MethodPost, "/order", ServiceHandler(s.order))
		r.Method(http.MethodPost, "/paid", ServiceHandler(s.paid))
	})
}

func (s *Service) getSessionVar(r *http.Request, name string) (any, error) {

	sessionName := s.Config.Session.Name
	session, err := s.SessionStore.Get(r, sessionName)
	if err != nil {
		return nil, fmt.Errorf("error fetching session %s: %w", sessionName, err)
	}
	return session.Values[name], nil
}

func (s *Service) setSessionVar(r *http.Request, w http.ResponseWriter, name string, value any) error {

	sessionName := s.Config.Session.Name
	session, err := s.SessionStore.Get(r, sessionName)
	if err != nil {
		return fmt.Errorf("error fetching session %s: %w", sessionName, err)
	}

	session.Values[name] = value
	return session.Save(r, w)
}
