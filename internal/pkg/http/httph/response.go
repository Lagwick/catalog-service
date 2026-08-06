package httph

import (
	"encoding/json"
	"errors"
	"net/http"
)

type httpCoder interface {
	error
	HTTPStatus() int
}

func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	ErrorApply(r.Context(), err)

	var hc httpCoder
	if errors.As(err, &hc) {
		status := hc.HTTPStatus()

		ErrorApplyStatusCode(r, status)
		sendError(w, status, hc)

		return
	}

	ErrorApplyStatusCode(r, http.StatusInternalServerError)
	sendError(w, http.StatusInternalServerError, err)
}
