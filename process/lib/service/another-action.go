package service

import (
	"net/http"
)

type anotherActionPageData struct {
	Title   string
	Message string
	CurrVal int
}

func (s *Service) anotherAction(w http.ResponseWriter, r *http.Request) error {

	data := anotherActionPageData{
		Title:   "Another Action",
		Message: "This is another Action Page",
	}

	return s.render(w, "index.go.html", data, http.StatusOK)
}
