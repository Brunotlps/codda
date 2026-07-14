package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const readinessTimeout = 2 * time.Second

// ReadinessChecker verifies whether dependencies required to serve traffic
// are reachable.
type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

// NewRouter builds the HTTP router for the order API, wiring h's methods to
// their routes. The returned handler is ready to serve.
func NewRouter(h *Handler, readinessCheckers ...ReadinessChecker) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		for _, checker := range readinessCheckers {
			if checker == nil {
				writeReadinessUnavailable(w)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
			err := checker.Ping(ctx)
			cancel()
			if err != nil {
				writeReadinessUnavailable(w)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", h.CreateOrder)
		r.Get("/", h.ListOrders)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.FindOrderByID)
			r.Post("/pay", h.MarkOrderAsPaid)
			r.Post("/cancel", h.MarkOrderAsCancelled)
			r.Post("/ship", h.MarkOrderAsShipped)
		})
	})

	return r
}

func writeReadinessUnavailable(w http.ResponseWriter) {
	writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
		Error: ErrorDetail{
			Code:    "service_unavailable",
			Message: "service is not ready",
		},
	})
}
