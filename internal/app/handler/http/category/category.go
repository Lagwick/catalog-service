package hcategory

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
	"github.com/Lagwick/catalog-service/internal/app/service"
	"github.com/Lagwick/catalog-service/internal/pkg/http/httph"
)

type handler struct {
	svcCategory service.Category
}

func NewHandler(svcCategory service.Category) rhandler.Category {
	return &handler{svcCategory: svcCategory}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestCategoryCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	if err := req.Validate(); err != nil {
		httph.SendError(w, http.StatusBadRequest, err)
		return
	}

	category, err := h.svcCategory.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrAlreadyExists):
			httph.SendError(w, http.StatusBadRequest, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}
	resp := entity.ResponseCategory{
		GUID:      category.GUID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusCreated, resp)
}

func (h *handler) GetByGUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	guid, err := uuid.Parse(vars["guid"])
	if err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	category, err := h.svcCategory.GetByGUID(r.Context(), guid)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			httph.SendError(w, http.StatusNotFound, err)
			return
		}
		httph.SendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := entity.ResponseCategory{
		GUID:      category.GUID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusOK, resp)
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	guid, err := uuid.Parse(vars["guid"])
	if err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	var req entity.RequestCategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	if err := req.Validate(); err != nil {
		httph.SendError(w, http.StatusBadRequest, err)
		return
	}

	category, err := h.svcCategory.Update(r.Context(), guid, req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			httph.SendError(w, http.StatusNotFound, err)
		case errors.Is(err, entity.ErrAlreadyExists):
			httph.SendError(w, http.StatusBadRequest, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := entity.ResponseCategory{
		GUID:      category.GUID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusOK, resp)
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	guid, err := uuid.Parse(vars["guid"])
	if err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	err = h.svcCategory.Delete(r.Context(), guid)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			httph.SendError(w, http.StatusNotFound, err)
		case errors.Is(err, entity.ErrCategoryHasProducts):
			httph.SendError(w, http.StatusBadRequest, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}

	httph.SendEmpty(w, http.StatusNoContent)
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svcCategory.List(r.Context())
	if err != nil {
		httph.SendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := make([]entity.ResponseCategory, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, entity.ResponseCategory{
			GUID:      c.GUID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	httph.SendJSON(w, http.StatusOK, resp)
}
