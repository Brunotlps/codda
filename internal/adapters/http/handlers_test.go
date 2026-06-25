package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	coddaHTTP "github.com/Brunotlps/codda/internal/adapters/http"
	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

// testEnv bundles a running test server with the repository backing it, so
// tests can both hit the HTTP API and inspect persisted state directly.
type testEnv struct {
	server *httptest.Server
	repo   *memory.OrderRepository
}

// setupTestServer wires a full Handler/Router stack backed by a fresh
// memory.OrderRepository and serves it over httptest.NewServer. The server
// is closed automatically when the test completes.
func setupTestServer(t *testing.T) *testEnv {
	t.Helper()

	repo := memory.NewOrderRepository()

	createOrder := application.NewCreateOrderUseCase(repo)
	findOrder := application.NewFindOrderByIDUseCase(repo)
	listOrders := application.NewListOrdersUseCase(repo)
	markPaid := application.NewMarkOrderAsPaidUseCase(repo)
	markCancelled := application.NewMarkOrderAsCancelledUseCase(repo)
	markShipped := application.NewMarkOrderAsShippedUseCase(repo)

	handler := coddaHTTP.NewHandler(createOrder, findOrder, listOrders, markPaid, markCancelled, markShipped)
	router := coddaHTTP.NewRouter(handler)
	server := httptest.NewServer(router)

	t.Cleanup(server.Close)

	return &testEnv{server: server, repo: repo}
}

// seedOrder builds a domain.Order with a single item and saves it directly
// into repo, for test setup.
func seedOrder(t *testing.T, repo *memory.OrderRepository) *domain.Order {
	t.Helper()

	price, err := domain.NewMoney(1000)
	if err != nil {
		t.Fatalf("NewMoney(...) failed during test setup: %v", err)
	}

	item, err := domain.NewOrderItem("p1", "Widget", price, 2)
	if err != nil {
		t.Fatalf("NewOrderItem(...) failed during test setup: %v", err)
	}

	order, err := domain.NewOrder([]domain.OrderItem{item})
	if err != nil {
		t.Fatalf("NewOrder(...) failed during test setup: %v", err)
	}

	if err := repo.Save(context.Background(), order); err != nil {
		t.Fatalf("Save(...) failed during test setup: %v", err)
	}

	return order
}

// postJSON encodes body as JSON and POSTs it to url, failing the test
// immediately if encoding or the request itself fails.
func postJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatalf("encode body: %v", err)
	}

	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	return resp
}

// decodeJSON decodes resp's body into target and closes the body. Callers
// must not close resp.Body themselves.
func decodeJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode body: %v", err)
	}
}

func TestHealthcheck(t *testing.T) {
	env := setupTestServer(t)

	resp, err := http.Get(env.server.URL + "/health")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCreateOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTestServer(t)

		reqBody := coddaHTTP.CreateOrderRequest{
			Items: []coddaHTTP.CreateOrderItemRequest{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 2},
			},
		}

		resp := postJSON(t, env.server.URL+"/orders", reqBody)
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}

		var created coddaHTTP.CreateOrderResponse
		decodeJSON(t, resp, &created)
		if created.ID == "" {
			t.Errorf("ID = %q, want non-empty", created.ID)
		}

		order, err := env.repo.FindByID(context.Background(), domain.OrderID(created.ID))
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got := order.Items(); len(got) != 1 || got[0].ProductID() != "p1" {
			t.Errorf("Items() = %v, want a single item with product id %q", got, "p1")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		env := setupTestServer(t)

		resp, err := http.Post(env.server.URL+"/orders", "application/json", strings.NewReader("{not valid json"))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			resp.Body.Close()
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}

		var errResp coddaHTTP.ErrorResponse
		decodeJSON(t, resp, &errResp)
		if errResp.Error.Code != "validation_error" {
			t.Errorf("Error.Code = %q, want %q", errResp.Error.Code, "validation_error")
		}
	})

	t.Run("invalid invariant", func(t *testing.T) {
		env := setupTestServer(t)

		reqBody := coddaHTTP.CreateOrderRequest{Items: nil}

		resp := postJSON(t, env.server.URL+"/orders", reqBody)
		if resp.StatusCode != http.StatusBadRequest {
			resp.Body.Close()
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}

		var errResp coddaHTTP.ErrorResponse
		decodeJSON(t, resp, &errResp)
		if errResp.Error.Code != "validation_error" {
			t.Errorf("Error.Code = %q, want %q", errResp.Error.Code, "validation_error")
		}
	})
}

func TestFindOrderByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		env := setupTestServer(t)
		order := seedOrder(t, env.repo)

		resp, err := http.Get(env.server.URL + "/orders/" + string(order.ID()))
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var got coddaHTTP.OrderResponse
		decodeJSON(t, resp, &got)

		if got.ID != string(order.ID()) {
			t.Errorf("ID = %q, want %q", got.ID, order.ID())
		}
		if got.TotalCents != order.Total().Cents() {
			t.Errorf("TotalCents = %d, want %d", got.TotalCents, order.Total().Cents())
		}
		if len(got.Items) != 1 {
			t.Errorf("Items = %v, want 1 item", got.Items)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := setupTestServer(t)

		resp, err := http.Get(env.server.URL + "/orders/missing-id")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			resp.Body.Close()
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}

		var errResp coddaHTTP.ErrorResponse
		decodeJSON(t, resp, &errResp)
		if errResp.Error.Code != "order_not_found" {
			t.Errorf("Error.Code = %q, want %q", errResp.Error.Code, "order_not_found")
		}
	})
}

func TestListOrders(t *testing.T) {
	env := setupTestServer(t)
	seedOrder(t, env.repo)
	seedOrder(t, env.repo)

	resp, err := http.Get(env.server.URL + "/orders")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var got coddaHTTP.ListOrdersResponse
	decodeJSON(t, resp, &got)

	if len(got.Orders) != 2 {
		t.Errorf("Orders = %v, want 2 items", got.Orders)
	}
	if got.HasMore {
		t.Errorf("HasMore = true, want false")
	}
}

func TestMarkOrderAsPaid(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTestServer(t)
		order := seedOrder(t, env.repo)

		resp, err := http.Post(env.server.URL+"/orders/"+string(order.ID())+"/pay", "application/json", nil)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}

		got, err := env.repo.FindByID(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got.Status() != domain.StatusPaid {
			t.Errorf("Status() = %v, want %v", got.Status(), domain.StatusPaid)
		}
	})

	t.Run("invalid transition", func(t *testing.T) {
		env := setupTestServer(t)
		order := seedOrder(t, env.repo)
		if err := order.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() setup failed: %v", err)
		}
		if err := env.repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) setup failed: %v", err)
		}

		resp, err := http.Post(env.server.URL+"/orders/"+string(order.ID())+"/pay", "application/json", nil)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusConflict)
		}

		var errResp coddaHTTP.ErrorResponse
		decodeJSON(t, resp, &errResp)
		if errResp.Error.Code != "invalid_status_transition" {
			t.Errorf("Error.Code = %q, want %q", errResp.Error.Code, "invalid_status_transition")
		}
	})
}
