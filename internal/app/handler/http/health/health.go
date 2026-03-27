package rhealth

import (
	"net/http"

	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
)

type handler struct{}

func NewHandler() rhandler.Health {
	return &handler{}
}

func (h *handler) LastCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
