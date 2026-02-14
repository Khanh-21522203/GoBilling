package http

import (
	"encoding/json"
	"net/http"

	"gobilling/internal/platform/errors"
)

type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	RequestID  string      `json:"request_id"`
	HasMore    bool        `json:"has_more,omitempty"`
	NextCursor string      `json:"next_cursor,omitempty"`
	Extra      interface{} `json:"extra,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code      string              `json:"code"`
	Message   string              `json:"message"`
	Details   []ValidationError   `json:"details,omitempty"`
	RequestID string              `json:"request_id"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
		Meta: meta,
	}

	json.NewEncoder(w).Encode(response)
}

func RespondError(w http.ResponseWriter, statusCode int, err error, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var e *errors.Error
	var code, message string

	if errors.Is(err, e) {
		code = e.Code
		message = e.Message
	} else {
		code = "INTERNAL_ERROR"
		message = "An internal error occurred"
	}

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func ErrorToHTTPStatus(err error) int {
	var e *errors.Error
	if !errors.Is(err, e) {
		return http.StatusInternalServerError
	}

	switch e.Code {
	case "NOT_FOUND", "CUSTOMER_NOT_FOUND", "PRODUCT_NOT_FOUND", "PLAN_NOT_FOUND",
		"SUBSCRIPTION_NOT_FOUND", "INVOICE_NOT_FOUND", "PAYMENT_NOT_FOUND":
		return http.StatusNotFound
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "CONFLICT", "SUBSCRIPTION_ALREADY_CANCELED", "INVOICE_ALREADY_PAID",
		"REFUND_ALREADY_EXISTS", "IDEMPOTENCY_CONFLICT":
		return http.StatusConflict
	case "INVALID_STATE", "CUSTOMER_DELETED", "PLAN_INACTIVE", "SUBSCRIPTION_NOT_ACTIVE",
		"INVOICE_NOT_OPEN", "PAYMENT_FAILED", "CUSTOMER_HAS_ACTIVE_SUBSCRIPTIONS":
		return http.StatusUnprocessableEntity
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests
	case "REQUEST_IN_PROGRESS":
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
