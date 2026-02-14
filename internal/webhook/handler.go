package webhook

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
	r.Route("/webhook_endpoints", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type WebhookEndpointResponse struct {
	ID        string            `json:"id"`
	URL       string            `json:"url"`
	Secret    string            `json:"secret,omitempty"`
	Events    []string          `json:"events"`
	Active    bool              `json:"active"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	endpoint, err := h.service.CreateEndpoint(r.Context(), req.URL, req.Events)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, toResponse(endpoint, true), nil)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	endpoint, err := h.service.GetEndpoint(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(endpoint, false), nil)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)

	endpoints, err := h.service.ListEndpoints(r.Context(), nil, limit+1)
	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(endpoints) > limit
	if hasMore {
		endpoints = endpoints[:limit]
	}

	var lastID string
	if len(endpoints) > 0 {
		lastID = endpoints[len(endpoints)-1].ID
	}

	responses := make([]WebhookEndpointResponse, len(endpoints))
	for i, e := range endpoints {
		responses[i] = toResponse(e, false)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		URL    *string  `json:"url,omitempty"`
		Events []string `json:"events,omitempty"`
		Active *bool    `json:"active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	endpoint, err := h.service.UpdateEndpoint(r.Context(), id, req.URL, req.Events, req.Active)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(endpoint, false), nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteEndpoint(r.Context(), id); err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toResponse(e *WebhookEndpoint, includeSecret bool) WebhookEndpointResponse {
	resp := WebhookEndpointResponse{
		ID:        e.ID,
		URL:       e.URL,
		Events:    e.Events,
		Active:    e.Active,
		Metadata:  e.Metadata,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: e.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if includeSecret {
		resp.Secret = e.Secret
	}

	return resp
}
