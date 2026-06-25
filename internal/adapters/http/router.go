package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds the HTTP router for the order API, wiring h's methods to
// their routes. The returned handler is ready to serve.
func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
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
