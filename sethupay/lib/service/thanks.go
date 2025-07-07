package service

import (
	"net/http"
)

func (s *Service) thanks(w http.ResponseWriter, r *http.Request) error {

	s.render(w, "thank-you.go.html", nil, http.StatusOK)
	return nil
}
