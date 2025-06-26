package service

import (
	"log/slog"
	"net/http"
	"sethupay/lib/config"

	"github.com/go-chi/httplog/v3"
)

func newSlogger(cfg *config.Config, logger *slog.Logger) func(http.Handler) http.Handler {

	return httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaOTEL.Concise(!cfg.InProduction),
		RecoverPanics: true,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},
		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},
		LogRequestBody:     wantVerboseLogs,
		LogResponseBody:    wantVerboseLogs,
	})
}

func wantVerboseLogs(r *http.Request) bool {
	return r.Header.Get("Debug") == "reveal-body-logs"
}
