package hcategory

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
	srv service.Category
}

func NewHandler(srv service.Category) rhandler.Category {
	return &handler{srv: srv}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestCategoryCreate
	if err := binding.ScanAndValidateJSON(r, &req); err != nil {
		httph.HandleError(w, r, err)
		return
	}

	category, err := h.srv.Create(r.Context(), req)
	if err != nil {
		httph.HandleError(w, r, err)
		return
	}

	resp := entity.ResponseCategoryCreate{
		GUID:      category.GUID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
	}

	httph.SendJSON(w, http.StatusCreated, resp)
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	guid, err := uuid.Parse(mux.Vars(r)["guid"])
	if err != nil {
		httph.HandleError(w, r, entity.ErrIncorrectParameters)
		return
	}

	var req entity.RequestCategoryUpdate
	if err := binding.ScanAndValidateJSON(r, &req); err != nil {
		httph.HandleError(w, r, err)
		return
	}

	category, err := h.srv.Update(r.Context(), guid, req)
	if err != nil {
		httph.HandleError(w, r, err)
		return
	}

	resp := entity.ResponseCategoryUpdate{
		GUID:      category.GUID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
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
	categories, err := h.srv.List(r.Context())
	if err != nil {
		httph.HandleError(w, r, err)
		return
	}

	resp := entity.ResponseCategoryList{
		Data: make([]entity.ResponseCategoryListItem, 0, len(categories)),
	}
	for _, c := range categories {
		resp.Data = append(resp.Data, entity.ResponseCategoryListItem{
			GUID:      c.GUID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	httph.SendJSON(w, http.StatusOK, resp)
}
