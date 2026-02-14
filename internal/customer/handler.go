package customer

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	
	httpPkg "gobilling/internal/platform/http"
	"gobilling/internal/pkg/pagination"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/customers", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/suspend", h.Suspend)
		r.Post("/{id}/reactivate", h.Reactivate)
	})
}

type CreateCustomerRequest struct {
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	ExternalID *string           `json:"external_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type CustomerResponse struct {
	ID         string            `json:"id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	ExternalID *string           `json:"external_id,omitempty"`
	Status     string            `json:"status"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	customer, err := h.service.Create(r.Context(), CreateRequest{
		Email:      req.Email,
		Name:       req.Name,
		ExternalID: req.ExternalID,
		Metadata:   req.Metadata,
	})

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, toResponse(customer), nil)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	customer, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(customer), nil)
}

type UpdateCustomerRequest struct {
	Name       *string           `json:"name,omitempty"`
	Email      *string           `json:"email,omitempty"`
	ExternalID *string           `json:"external_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	customer, err := h.service.Update(r.Context(), id, UpdateRequest{
		Name:       req.Name,
		Email:      req.Email,
		ExternalID: req.ExternalID,
		Metadata:   req.Metadata,
	})

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(customer), nil)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)
	startingAfter := r.URL.Query().Get("starting_after")

	customers, err := h.service.List(r.Context(), ListOptions{
		Limit:         limit + 1,
		StartingAfter: startingAfter,
	})

	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(customers) > limit
	if hasMore {
		customers = customers[:limit]
	}

	var lastID string
	if len(customers) > 0 {
		lastID = customers[len(customers)-1].ID
	}

	responses := make([]CustomerResponse, len(customers))
	for i, c := range customers {
		responses[i] = toResponse(c)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Suspend(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	customer, err := h.service.Suspend(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(customer), nil)
}

func (h *Handler) Reactivate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	customer, err := h.service.Reactivate(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(customer), nil)
}

func toResponse(c *Customer) CustomerResponse {
	return CustomerResponse{
		ID:         c.ID,
		Email:      c.Email,
		Name:       c.Name,
		ExternalID: c.ExternalID,
		Status:     string(c.Status),
		Metadata:   c.Metadata,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
