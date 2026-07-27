package hproduct

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
	"github.com/Lagwick/catalog-service/internal/app/service"
	"github.com/Lagwick/catalog-service/internal/pkg/http/binding"
	"github.com/Lagwick/catalog-service/internal/pkg/http/httph"
)

type handler struct {
	srv service.Product
}

func NewHandler(srv service.Product) rhandler.Product {
	return &handler{srv: srv}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestProductCreate
	if err := binding.ScanAndValidateJSON(r, &req); err != nil {
		httph.HandleError(w, r, err)
		return
	}

	product, err := h.srv.Create(r.Context(), req)
	if err != nil {
		httph.HandleError(w, r, err)
		return
	}

	resp := entity.ResponseProductCreate{
		GUID:         product.GUID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		CategoryGUID: product.CategoryGUID,
		CreatedAt:    product.CreatedAt,
	}

	httph.SendJSON(w, http.StatusCreated, resp)
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	guid, err := uuid.Parse(mux.Vars(r)["guid"])
	if err != nil {
		httph.HandleError(w, r, entity.ErrIncorrectParameters)
		return
	}

	var req entity.RequestProductUpdate
	if err := binding.ScanAndValidateJSON(r, &req); err != nil {
		httph.HandleError(w, r, err)
		return
	}

	product, err := h.srv.Update(r.Context(), guid, req)
	if err != nil {
		httph.HandleError(w, r, err)
		return
	}

	resp := entity.ResponseProductUpdate{
		GUID:         product.GUID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		CategoryGUID: product.CategoryGUID,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusOK, resp)
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	guid, err := uuid.Parse(mux.Vars(r)["guid"])
	if err != nil {
		httph.HandleError(w, r, entity.ErrIncorrectParameters)
		return
	}

	if err := h.srv.Delete(r.Context(), guid); err != nil {
		httph.HandleError(w, r, err)
		return
	}

	httph.SendEmpty(w, http.StatusOK)
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestProductList
	if r.Body != nil {
		if err := binding.ScanAndValidateJSON(r, &req); err != nil {
			httph.HandleError(w, r, err)
			return
		}
	}

	products, err := h.srv.List(r.Context(), req)
	if err != nil {
		httph.HandleError(w, r, err)
		return
	}

	resp := entity.ResponseProductList{
		Data: make([]entity.ResponseProductListItem, 0, len(products)),
	}
	for _, p := range products {
		resp.Data = append(resp.Data, entity.ResponseProductListItem{
			GUID:         p.GUID,
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			CategoryGUID: p.CategoryGUID,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
		})
	}

	httph.SendJSON(w, http.StatusOK, resp)
}
