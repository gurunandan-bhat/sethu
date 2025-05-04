package service

import (
	"fmt"
	"net/http"
)

type IndexPageData struct {
	Title   string
	Message string
	CurrVal int
}

func (s *Service) index(w http.ResponseWriter, r *http.Request) error {

	iCurrVal := int(1)
	data := IndexPageData{
		Title:   "Home",
		Message: "This is the Home Page",
		CurrVal: iCurrVal,
	}

	currVal := int(1)
	if err := s.setSessionVar(r, w, "currval", currVal); err != nil {
		return fmt.Errorf("error saving key currval to %d: %w", currVal, err)
	}

	return s.render(w, "index.go.html", data, http.StatusOK)
}
