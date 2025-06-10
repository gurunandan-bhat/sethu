package service

import (
	"fmt"
	"net/http"
)

type AboutPageData struct {
	Title   string
	Message string
	CurrVal int
}

func (s *Service) about(w http.ResponseWriter, r *http.Request) error {

	currval, err := s.getSessionVar(r, "currval")
	if err != nil || currval == nil {
		fmt.Println("session values has no key currval. setting it to 0")
		currval = 0
	}
	currVal, ok := currval.(int)
	if !ok {
		return fmt.Errorf("currval does not have the right integer type")
	}

	if err := s.setSessionVar(r, w, "currval", currVal+1); err != nil {
		return fmt.Errorf("error incrementing session value currval")
	}

	data := AboutPageData{
		Title:   "About",
		Message: "This is the About Page",
		CurrVal: currVal,
	}

	return s.render(w, "index.go.html", data, http.StatusOK)
}
