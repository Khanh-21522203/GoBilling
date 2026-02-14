package subscription

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
	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Post("/{id}/cancel", h.Cancel)
		r.Patch("/{id}", h.UpdatePlan)
	})
}

type CreateSubscriptionRequest struct {
	CustomerID string            `json:"customer_id"`
	PlanID     string            `json:"plan_id"`
	Quantity   int               `json:"quantity,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type SubscriptionResponse struct {
	ID                 string            `json:"id"`
	CustomerID         string            `json:"customer_id"`
	PlanID             string            `json:"plan_id"`
	Status             string            `json:"status"`
	Quantity           int               `json:"quantity"`
	CurrentPeriodStart string            `json:"current_period_start"`
	CurrentPeriodEnd   string            `json:"current_period_end"`
	TrialStart         *string           `json:"trial_start,omitempty"`
	TrialEnd           *string           `json:"trial_end,omitempty"`
	CancelAtPeriodEnd  bool              `json:"cancel_at_period_end"`
	CanceledAt         *string           `json:"canceled_at,omitempty"`
	Metadata           map[string]string `json:"metadata"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	sub, err := h.service.Create(r.Context(), CreateRequest{
		CustomerID: req.CustomerID,
		PlanID:     req.PlanID,
		Quantity:   req.Quantity,
		Metadata:   req.Metadata,
	})

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, toResponse(sub), nil)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(sub), nil)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)
	startingAfter := r.URL.Query().Get("starting_after")
	customerID := r.URL.Query().Get("customer_id")

	var customerIDPtr *string
	if customerID != "" {
		customerIDPtr = &customerID
	}

	subs, err := h.service.List(r.Context(), ListOptions{
		Limit:         limit + 1,
		StartingAfter: startingAfter,
		CustomerID:    customerIDPtr,
	})

	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(subs) > limit
	if hasMore {
		subs = subs[:limit]
	}

	var lastID string
	if len(subs) > 0 {
		lastID = subs[len(subs)-1].ID
	}

	responses := make([]SubscriptionResponse, len(subs))
	for i, s := range subs {
		responses[i] = toResponse(s)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	sub, err := h.service.Cancel(r.Context(), id, CancelRequest{
		CancelAtPeriodEnd: req.CancelAtPeriodEnd,
	})

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(sub), nil)
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		PlanID string `json:"plan_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	sub, err := h.service.UpdatePlan(r.Context(), id, UpdatePlanRequest{
		NewPlanID: req.PlanID,
	})

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(sub), nil)
}

func toResponse(s *Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:                 s.ID,
		CustomerID:         s.CustomerID,
		PlanID:             s.PlanID,
		Status:             string(s.Status),
		Quantity:           s.Quantity,
		CurrentPeriodStart: s.CurrentPeriodStart.Format("2006-01-02T15:04:05Z07:00"),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z07:00"),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Metadata:           s.Metadata,
		CreatedAt:          s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if s.TrialStart != nil {
		ts := s.TrialStart.Format("2006-01-02T15:04:05Z07:00")
		resp.TrialStart = &ts
	}

	if s.TrialEnd != nil {
		te := s.TrialEnd.Format("2006-01-02T15:04:05Z07:00")
		resp.TrialEnd = &te
	}

	if s.CanceledAt != nil {
		ca := s.CanceledAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CanceledAt = &ca
	}

	return resp
}
