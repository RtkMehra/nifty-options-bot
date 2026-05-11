package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DhanClient struct {
	clientID    string
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

func NewDhanClient(clientID, accessToken string) *DhanClient {
	return &DhanClient{
		clientID:    clientID,
		accessToken: accessToken,
		baseURL:     "https://api.dhan.co",
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

type OrderRequest struct {
	DhanClientID    string  `json:"dhanClientId"`
	TransactionType string  `json:"transactionType"`
	ExchangeSegment string  `json:"exchangeSegment"`
	ProductType     string  `json:"productType"`
	OrderType       string  `json:"orderType"`
	Validity        string  `json:"validity"`
	TradingSymbol   string  `json:"tradingSymbol"`
	SecurityID      string  `json:"securityId"`
	Quantity        int     `json:"quantity"`
	Price           float64 `json:"price"`
}

type OrderResponse struct {
	OrderID     string `json:"orderId"`
	OrderStatus string `json:"orderStatus"`
}

func (d *DhanClient) PlaceOrder(req OrderRequest) (OrderResponse, error) {
	req.DhanClientID = d.clientID

	body, err := json.Marshal(req)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("marshal order: %w", err)
	}

	httpReq, err := http.NewRequest("POST", d.baseURL+"/orders", bytes.NewReader(body))
	if err != nil {
		return OrderResponse{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("access-token", d.accessToken)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("place order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return OrderResponse{}, fmt.Errorf("order failed: status=%d", resp.StatusCode)
	}

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return OrderResponse{}, fmt.Errorf("decode response: %w", err)
	}

	return orderResp, nil
}

func (d *DhanClient) CancelOrder(orderID string) error {
	httpReq, err := http.NewRequest("DELETE", d.baseURL+"/orders/"+orderID, nil)
	if err != nil {
		return fmt.Errorf("create cancel request: %w", err)
	}
	httpReq.Header.Set("access-token", d.accessToken)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel failed: status=%d", resp.StatusCode)
	}
	return nil
}

func (d *DhanClient) GetOrderStatus(orderID string) (string, error) {
	httpReq, err := http.NewRequest("GET", d.baseURL+"/orders/"+orderID, nil)
	if err != nil {
		return "", fmt.Errorf("create status request: %w", err)
	}
	httpReq.Header.Set("access-token", d.accessToken)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("get order status: %w", err)
	}
	defer resp.Body.Close()

	var statusResp struct {
		OrderStatus string `json:"orderStatus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return "", fmt.Errorf("decode status: %w", err)
	}

	return statusResp.OrderStatus, nil
}
