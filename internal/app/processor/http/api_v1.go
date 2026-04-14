package rprocessor

import (
	"net/http"

	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
	"github.com/gorilla/mux"
)

func v1RegCategoryHandler(r1 *mux.Router, h rhandler.Category) {
	reg(r1, http.MethodPost, "/category/create", http.HandlerFunc(h.Create))
	reg(r1, http.MethodGet, "/category/{guid}", http.HandlerFunc(h.GetByGUID))
	reg(r1, http.MethodPatch, "/category/{guid}", http.HandlerFunc(h.Update))
	reg(r1, http.MethodDelete, "/category/{guid}", http.HandlerFunc(h.Delete))
	reg(r1, http.MethodPost, "/category/list", http.HandlerFunc(h.List))
}

func v1RegProductHandler(r1 *mux.Router, h rhandler.Product) {
	reg(r1, http.MethodPost, "/product/create", http.HandlerFunc(h.Create))
	reg(r1, http.MethodGet, "/product/{guid}", http.HandlerFunc(h.GetByGUID))
	reg(r1, http.MethodPatch, "/product/{guid}", http.HandlerFunc(h.Update))
	reg(r1, http.MethodDelete, "/product/{guid}", http.HandlerFunc(h.Delete))
	reg(r1, http.MethodPost, "/product/list", http.HandlerFunc(h.List))
}
