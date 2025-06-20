package service

import (
	"fmt"
	"net/http"
)

type ServiceHandler func(w http.ResponseWriter, r *http.Request) error

func (h ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		errStr := fmt.Sprintf("%+v", err)
		http.Error(
			w,
			"from root handler: "+errStr,
			http.StatusBadRequest,
		)
		return
	}

	return
}
