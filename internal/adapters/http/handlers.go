package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

// errDecodeRequest is returned when a request body fails to decode as JSON.
var errDecodeRequest = errors.New("invalid request body")

// errMissingOrderID is returned when a request is missing the order ID URL
// parameter.
var errMissingOrderID = errors.New("order id is required")

// maxBodyBytes is the maximum size accepted for a request body, guarding
// against unbounded memory growth from oversized or malicious payloads.
const maxBodyBytes = 1 << 20 // 1 MiB

// Handler holds the use cases needed to serve the order HTTP API.
type Handler struct {
	createOrder   *application.CreateOrderUseCase
	findOrder     *application.FindOrderByIDUseCase
	listOrders    *application.ListOrdersUseCase
	markPaid      *application.MarkOrderAsPaidUseCase
	markCancelled *application.MarkOrderAsCancelledUseCase
	markShipped   *application.MarkOrderAsShippedUseCase
}

// NewHandler creates a Handler backed by the given use cases.
func NewHandler(
	createOrder *application.CreateOrderUseCase,
	findOrder *application.FindOrderByIDUseCase,
	listOrders *application.ListOrdersUseCase,
	markPaid *application.MarkOrderAsPaidUseCase,
	markCancelled *application.MarkOrderAsCancelledUseCase,
	markShipped *application.MarkOrderAsShippedUseCase,
) *Handler {
	return &Handler{
		createOrder:   createOrder,
		findOrder:     findOrder,
		listOrders:    listOrders,
		markPaid:      markPaid,
		markCancelled: markCancelled,
		markShipped:   markShipped,
	}
}

// CreateOrder handles POST /orders.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errDecodeRequest)
		return
	}

	input := application.CreateOrderInput{
		Items: make([]application.CreateOrderItemInput, len(req.Items)),
	}
	for i, item := range req.Items {
		input.Items[i] = application.CreateOrderItemInput{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			PriceCents:  item.PriceCents,
			Quantity:    item.Quantity,
		}
	}

	id, err := h.createOrder.Execute(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, CreateOrderResponse{ID: string(id)})
}

// FindOrderByID handles GET /orders/{id}.
func (h *Handler) FindOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, errMissingOrderID)
		return
	}

	order, err := h.findOrder.Execute(r.Context(), domain.OrderID(id))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponse(order))
}

// ListOrders handles GET /orders.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var filters application.ListOrdersFilters
	if statusStr := q.Get("status"); statusStr != "" {
		status := domain.OrderStatus(statusStr)
		filters.Status = &status
	}

	var pagination application.Pagination
	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			pagination.Limit = limit
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			pagination.Offset = offset
		}
	}

	resumes, hasMore, err := h.listOrders.Execute(r.Context(), filters, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := ListOrdersResponse{
		Orders:  make([]OrderResumeResponse, len(resumes)),
		HasMore: hasMore,
	}
	for i, resume := range resumes {
		resp.Orders[i] = toOrderResumeResponse(resume)
	}

	writeJSON(w, http.StatusOK, resp)
}

// MarkOrderAsPaid handles POST /orders/{id}/pay.
func (h *Handler) MarkOrderAsPaid(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, errMissingOrderID)
		return
	}

	if err := h.markPaid.Execute(r.Context(), domain.OrderID(id)); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

// MarkOrderAsCancelled handles POST /orders/{id}/cancel.
func (h *Handler) MarkOrderAsCancelled(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, errMissingOrderID)
		return
	}

	if err := h.markCancelled.Execute(r.Context(), domain.OrderID(id)); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

// MarkOrderAsShipped handles POST /orders/{id}/ship.
func (h *Handler) MarkOrderAsShipped(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, errMissingOrderID)
		return
	}

	if err := h.markShipped.Execute(r.Context(), domain.OrderID(id)); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

// writeJSON writes body as a JSON response with the given status code. If
// body is nil, only the status code is written.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// writeError writes err as an ErrorResponse, with the status code and error
// code determined by httpStatusForError. If err is or wraps
// context.Canceled, writeError writes nothing: the client has already
// disconnected, so there is no one to receive the response.
func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	status, code, message := httpStatusForError(err)
	writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{Code: code, Message: message},
	})
}

// httpStatusForError maps err to an HTTP status code, a stable error code,
// and a human-readable message.
func httpStatusForError(err error) (int, string, string) {
	switch {
	case errors.Is(err, application.ErrOrderNotFound):
		return http.StatusNotFound, "order_not_found", "order not found"
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return http.StatusConflict, "invalid_status_transition", "order cannot transition to this status"
	case errors.Is(err, domain.ErrOrderRequiresItems),
		errors.Is(err, domain.ErrEmptyProductID),
		errors.Is(err, domain.ErrEmptyProductName),
		errors.Is(err, domain.ErrProductNameTooLong),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrNegativeMoney),
		errors.Is(err, errDecodeRequest),
		errors.Is(err, errMissingOrderID):
		return http.StatusBadRequest, "validation_error", err.Error()
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusServiceUnavailable, "service_unavailable", "request could not be completed in time"
	default:
		return http.StatusInternalServerError, "internal_error", "internal server error"
	}
}

// toOrderResponse maps a domain.Order to its full HTTP representation.
func toOrderResponse(o *domain.Order) OrderResponse {
	items := o.Items()
	itemResponses := make([]OrderItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = toOrderItemResponse(item)
	}

	return OrderResponse{
		ID:         string(o.ID()),
		Status:     string(o.Status()),
		Items:      itemResponses,
		TotalCents: o.Total().Cents(),
		CreatedAt:  o.CreatedAt(),
	}
}

// toOrderItemResponse maps a domain.OrderItem to its HTTP representation.
func toOrderItemResponse(i domain.OrderItem) OrderItemResponse {
	return OrderItemResponse{
		ProductID:   i.ProductID(),
		ProductName: i.ProductName(),
		PriceCents:  i.Price().Cents(),
		Quantity:    i.Quantity(),
	}
}

// toOrderResumeResponse maps an application.OrderResume to its HTTP
// representation.
func toOrderResumeResponse(resume application.OrderResume) OrderResumeResponse {
	return OrderResumeResponse{
		ID:         string(resume.ID),
		Status:     string(resume.Status),
		TotalCents: resume.Total.Cents(),
		CreatedAt:  resume.CreatedAt,
	}
}
