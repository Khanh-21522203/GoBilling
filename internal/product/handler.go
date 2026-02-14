package product

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
	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.ListProducts)
		r.Get("/{id}", h.GetProduct)
		r.Patch("/{id}", h.UpdateProduct)
		r.Post("/{id}/archive", h.ArchiveProduct)
	})

	r.Route("/plans", func(r chi.Router) {
		r.Post("/", h.CreatePlan)
		r.Get("/", h.ListPlans)
		r.Get("/{id}", h.GetPlan)
		r.Post("/{id}/archive", h.ArchivePlan)
	})
}

type CreateProductRequest struct {
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ProductResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Active      bool              `json:"active"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req.Name, req.Description, req.Metadata)

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, toProductResponse(product), nil)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toProductResponse(product), nil)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Name        *string           `json:"name,omitempty"`
		Description *string           `json:"description,omitempty"`
		Metadata    map[string]string `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, UpdateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Metadata:    req.Metadata,
	})

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toProductResponse(product), nil)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)
	startingAfter := r.URL.Query().Get("starting_after")

	products, err := h.service.ListProducts(r.Context(), ProductListOptions{
		Limit:         limit + 1,
		StartingAfter: startingAfter,
	})

	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	var lastID string
	if len(products) > 0 {
		lastID = products[len(products)-1].ID
	}

	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = toProductResponse(p)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) ArchiveProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	product, err := h.service.ArchiveProduct(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toProductResponse(product), nil)
}

type CreatePlanRequest struct {
	ProductID            string        `json:"product_id"`
	Name                 string        `json:"name"`
	Description          *string       `json:"description,omitempty"`
	PricingType          string        `json:"pricing_type"`
	Amount               int64         `json:"amount,omitempty"`
	Currency             string        `json:"currency"`
	BillingInterval      string        `json:"billing_interval"`
	BillingIntervalCount int           `json:"billing_interval_count,omitempty"`
	TrialPeriodDays      int           `json:"trial_period_days,omitempty"`
	Tiers                []PricingTier `json:"tiers,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

type PlanResponse struct {
	ID                   string            `json:"id"`
	ProductID            string            `json:"product_id"`
	Name                 string            `json:"name"`
	Description          *string           `json:"description,omitempty"`
	PricingType          string            `json:"pricing_type"`
	Amount               int64             `json:"amount,omitempty"`
	Currency             string            `json:"currency"`
	BillingInterval      string            `json:"billing_interval"`
	BillingIntervalCount int               `json:"billing_interval_count"`
	TrialPeriodDays      int               `json:"trial_period_days"`
	Tiers                []PricingTier     `json:"tiers,omitempty"`
	Active               bool              `json:"active"`
	Metadata             map[string]string `json:"metadata"`
	CreatedAt            string            `json:"created_at"`
	UpdatedAt            string            `json:"updated_at"`
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	if req.BillingIntervalCount == 0 {
		req.BillingIntervalCount = 1
	}

	plan, err := h.service.CreatePlan(
		r.Context(),
		req.ProductID,
		req.Name,
		req.Description,
		PricingType(req.PricingType),
		req.Amount,
		req.Currency,
		BillingInterval(req.BillingInterval),
		req.BillingIntervalCount,
		req.TrialPeriodDays,
		req.Tiers,
		req.Metadata,
	)

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, toPlanResponse(plan), nil)
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	plan, err := h.service.GetPlan(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toPlanResponse(plan), nil)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)
	startingAfter := r.URL.Query().Get("starting_after")
	productID := r.URL.Query().Get("product_id")

	var productIDPtr *string
	if productID != "" {
		productIDPtr = &productID
	}

	plans, err := h.service.ListPlans(r.Context(), PlanListOptions{
		Limit:         limit + 1,
		StartingAfter: startingAfter,
		ProductID:     productIDPtr,
	})

	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(plans) > limit
	if hasMore {
		plans = plans[:limit]
	}

	var lastID string
	if len(plans) > 0 {
		lastID = plans[len(plans)-1].ID
	}

	responses := make([]PlanResponse, len(plans))
	for i, p := range plans {
		responses[i] = toPlanResponse(p)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) ArchivePlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	plan, err := h.service.ArchivePlan(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toPlanResponse(plan), nil)
}

func toProductResponse(p *Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Active:      p.Active,
		Metadata:    p.Metadata,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toPlanResponse(p *Plan) PlanResponse {
	return PlanResponse{
		ID:                   p.ID,
		ProductID:            p.ProductID,
		Name:                 p.Name,
		Description:          p.Description,
		PricingType:          string(p.PricingType),
		Amount:               p.Amount,
		Currency:             p.Currency,
		BillingInterval:      string(p.BillingInterval),
		BillingIntervalCount: p.BillingIntervalCount,
		TrialPeriodDays:      p.TrialPeriodDays,
		Tiers:                p.Tiers,
		Active:               p.Active,
		Metadata:             p.Metadata,
		CreatedAt:            p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
