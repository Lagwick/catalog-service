package httph

import (
	"net/http"
)

type Error struct {
	Message string `json:"error"`
}

func ErrorApply(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)

	errResp := Error{
		Message: message,
	}

	_ = EncodeJSON(w, errResp)

}
