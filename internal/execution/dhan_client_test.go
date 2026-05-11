package execution

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDhanClient(t *testing.T) {
	c := NewDhanClient("test", "token")
	if c == nil {
		t.Fatal("client is nil")
	}
	if c.clientID != "test" {
		t.Errorf("clientID = %v, want test", c.clientID)
	}
}

func TestPlaceOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access-token") != "test-token" {
			t.Error("missing access-token header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OrderResponse{OrderID: "123", OrderStatus: "PENDING"})
	}))
	defer server.Close()

	c := NewDhanClient("client1", "test-token")
	c.baseURL = server.URL

	resp, err := c.PlaceOrder(OrderRequest{
		TransactionType: "BUY",
		ExchangeSegment: "NSE_FNO",
		TradingSymbol:   "NIFTY25JAN22500CE",
		Quantity:        75,
		Price:           185,
	})
	if err != nil {
		t.Fatalf("PlaceOrder() err = %v", err)
	}
	if resp.OrderID != "123" {
		t.Errorf("OrderID = %v, want 123", resp.OrderID)
	}
	if resp.OrderStatus != "PENDING" {
		t.Errorf("OrderStatus = %v, want PENDING", resp.OrderStatus)
	}
}

func TestPlaceOrder_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	c := NewDhanClient("client1", "test-token")
	c.baseURL = server.URL

	_, err := c.PlaceOrder(OrderRequest{
		TransactionType: "BUY",
		TradingSymbol:   "TEST",
		Quantity:        75,
	})
	if err == nil {
		t.Error("expected error for bad request, got nil")
	}
}

func TestCancelOrder(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			called = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewDhanClient("client1", "test-token")
	c.baseURL = server.URL

	err := c.CancelOrder("order-123")
	if err != nil {
		t.Fatalf("CancelOrder() err = %v", err)
	}
	if !called {
		t.Error("DELETE endpoint was not called")
	}
}

func TestGetOrderStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %v", r.Method)
		}
		json.NewEncoder(w).Encode(struct {
			OrderStatus string `json:"orderStatus"`
		}{OrderStatus: "TRADED"})
	}))
	defer server.Close()

	c := NewDhanClient("client1", "test-token")
	c.baseURL = server.URL

	status, err := c.GetOrderStatus("order-123")
	if err != nil {
		t.Fatalf("GetOrderStatus() err = %v", err)
	}
	if status != "TRADED" {
		t.Errorf("status = %v, want TRADED", status)
	}
}
