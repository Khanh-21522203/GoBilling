package invoice

import (
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
	r.Route("/invoices", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Post("/{id}/finalize", h.Finalize)
		r.Post("/{id}/void", h.Void)
	})
}

type InvoiceResponse struct {
	ID             string              `json:"id"`
	InvoiceNumber  string              `json:"invoice_number"`
	CustomerID     string              `json:"customer_id"`
	SubscriptionID *string             `json:"subscription_id,omitempty"`
	Status         string              `json:"status"`
	Currency       string              `json:"currency"`
	Subtotal       int64               `json:"subtotal"`
	DiscountAmount int64               `json:"discount_amount"`
	TaxAmount      int64               `json:"tax_amount"`
	Total          int64               `json:"total"`
	AmountPaid     int64               `json:"amount_paid"`
	AmountDue      int64               `json:"amount_due"`
	PeriodStart    *string             `json:"period_start,omitempty"`
	PeriodEnd      *string             `json:"period_end,omitempty"`
	DueDate        *string             `json:"due_date,omitempty"`
	PaidAt         *string             `json:"paid_at,omitempty"`
	LineItems      []LineItemResponse  `json:"line_items,omitempty"`
	Metadata       map[string]string   `json:"metadata"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

type LineItemResponse struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Quantity    int64             `json:"quantity"`
	UnitAmount  int64             `json:"unit_amount"`
	Amount      int64             `json:"amount"`
	PeriodStart *string           `json:"period_start,omitempty"`
	PeriodEnd   *string           `json:"period_end,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	inv, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(inv), nil)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)
	startingAfter := r.URL.Query().Get("starting_after")
	customerID := r.URL.Query().Get("customer_id")

	var customerIDPtr *string
	if customerID != "" {
		customerIDPtr = &customerID
	}

	invoices, err := h.service.List(r.Context(), ListOptions{
		Limit:         limit + 1,
		StartingAfter: startingAfter,
		CustomerID:    customerIDPtr,
	})

	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(invoices) > limit
	if hasMore {
		invoices = invoices[:limit]
	}

	var lastID string
	if len(invoices) > 0 {
		lastID = invoices[len(invoices)-1].ID
	}

	responses := make([]InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		responses[i] = toResponse(inv)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) Finalize(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	inv, err := h.service.Finalize(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(inv), nil)
}

func (h *Handler) Void(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	inv, err := h.service.Void(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toResponse(inv), nil)
}

func toResponse(inv *Invoice) InvoiceResponse {
	resp := InvoiceResponse{
		ID:             inv.ID,
		InvoiceNumber:  inv.InvoiceNumber,
		CustomerID:     inv.CustomerID,
		SubscriptionID: inv.SubscriptionID,
		Status:         string(inv.Status),
		Currency:       inv.Currency,
		Subtotal:       inv.Subtotal,
		DiscountAmount: inv.DiscountAmount,
		TaxAmount:      inv.TaxAmount,
		Total:          inv.Total,
		AmountPaid:     inv.AmountPaid,
		AmountDue:      inv.AmountDue,
		Metadata:       inv.Metadata,
		CreatedAt:      inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      inv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if inv.PeriodStart != nil {
		ps := inv.PeriodStart.Format("2006-01-02T15:04:05Z07:00")
		resp.PeriodStart = &ps
	}

	if inv.PeriodEnd != nil {
		pe := inv.PeriodEnd.Format("2006-01-02T15:04:05Z07:00")
		resp.PeriodEnd = &pe
	}

	if inv.DueDate != nil {
		dd := inv.DueDate.Format("2006-01-02T15:04:05Z07:00")
		resp.DueDate = &dd
	}

	if inv.PaidAt != nil {
		pa := inv.PaidAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PaidAt = &pa
	}

	if len(inv.LineItems) > 0 {
		resp.LineItems = make([]LineItemResponse, len(inv.LineItems))
		for i, item := range inv.LineItems {
			resp.LineItems[i] = toLineItemResponse(item)
		}
	}

	return resp
}

func toLineItemResponse(item *LineItem) LineItemResponse {
	resp := LineItemResponse{
		ID:          item.ID,
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitAmount:  item.UnitAmount,
		Amount:      item.Amount,
		Metadata:    item.Metadata,
	}

	if item.PeriodStart != nil {
		ps := item.PeriodStart.Format("2006-01-02T15:04:05Z07:00")
		resp.PeriodStart = &ps
	}

	if item.PeriodEnd != nil {
		pe := item.PeriodEnd.Format("2006-01-02T15:04:05Z07:00")
		resp.PeriodEnd = &pe
	}

	return resp
}
