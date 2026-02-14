package payment

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
	r.Route("/payments", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Post("/{id}/refund", h.Refund)
	})
}

type PaymentResponse struct {
	ID             string            `json:"id"`
	InvoiceID      string            `json:"invoice_id"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Status         string            `json:"status"`
	FailureCode    *string           `json:"failure_code,omitempty"`
	FailureMessage *string           `json:"failure_message,omitempty"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type RefundResponse struct {
	ID        string            `json:"id"`
	PaymentID string            `json:"payment_id"`
	Amount    int64             `json:"amount"`
	Currency  string            `json:"currency"`
	Status    string            `json:"status"`
	Reason    *string           `json:"reason,omitempty"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	payment, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toPaymentResponse(payment), nil)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := pagination.ValidateLimit(20)
	startingAfter := r.URL.Query().Get("starting_after")
	invoiceID := r.URL.Query().Get("invoice_id")

	var invoiceIDPtr *string
	if invoiceID != "" {
		invoiceIDPtr = &invoiceID
	}

	payments, err := h.service.List(r.Context(), ListOptions{
		Limit:         limit + 1,
		StartingAfter: startingAfter,
		InvoiceID:     invoiceIDPtr,
	})

	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	hasMore := len(payments) > limit
	if hasMore {
		payments = payments[:limit]
	}

	var lastID string
	if len(payments) > 0 {
		lastID = payments[len(payments)-1].ID
	}

	responses := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = toPaymentResponse(p)
	}

	meta := &httpPkg.Meta{
		HasMore:    hasMore,
		NextCursor: pagination.EncodeCursor(lastID),
	}

	httpPkg.RespondJSON(w, http.StatusOK, responses, meta)
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Reason string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	refund, err := h.service.Refund(r.Context(), id, req.Reason)

	if err != nil {
		status := httpPkg.ErrorToHTTPStatus(err)
		httpPkg.RespondError(w, status, err, "")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, toRefundResponse(refund), nil)
}

func toPaymentResponse(p *Payment) PaymentResponse {
	return PaymentResponse{
		ID:             p.ID,
		InvoiceID:      p.InvoiceID,
		Amount:         p.Amount,
		Currency:       p.Currency,
		Status:         string(p.Status),
		FailureCode:    p.FailureCode,
		FailureMessage: p.FailureMessage,
		Metadata:       p.Metadata,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toRefundResponse(r *Refund) RefundResponse {
	return RefundResponse{
		ID:        r.ID,
		PaymentID: r.PaymentID,
		Amount:    r.Amount,
		Currency:  r.Currency,
		Status:    string(r.Status),
		Reason:    r.Reason,
		Metadata:  r.Metadata,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
