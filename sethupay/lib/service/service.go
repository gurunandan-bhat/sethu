package service

import (
	"errors"
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

	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

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

	s.setRoutes(*cfg)

	return s, nil
}

func (s *Service) setRoutes(cfg config.Config) {

	fileServer := http.FileServer(http.Dir(cfg.HugoRoot + "/themes/sethu/assets"))
	s.Muxer.Get("/sethupay/assets/*", http.HandlerFunc(http.StripPrefix("/sethupay/assets", fileServer).ServeHTTP))

	s.Muxer.Route("/sethupay", func(r chi.Router) {

		r.Method(http.MethodPost, "/order", ServiceHandler(s.order))
		r.Method(http.MethodPost, "/paid", ServiceHandler(s.paid))
		r.Method(http.MethodGet, "/thanks", ServiceHandler(s.thanks))
	})
}

func (s *Service) RazorpaySecret() (config.Secret, error) {
	// Generate the expected signature
	cfg := s.Config
	key := cfg.RazorPay.Test
	if cfg.InProduction {
		key = cfg.RazorPay.Live
	}

	if key.KeyID == "" || key.KeySecret == "" {
		return key, errors.New("no valid keys found")
	}

	return key, nil
}
