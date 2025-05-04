package service

import (
	"fmt"
	"net/http"
)

type ServiceHandler func(w http.ResponseWriter, r *http.Request) error

func (h ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if err := h(w, r); err != nil {
		http.Error(
			w,
			fmt.Errorf("from root handler - error: %s", err).Error(),
			http.StatusBadRequest,
		)
		return
	}
}
