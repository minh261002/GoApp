package shipping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GHTKConfig represents GHTK API configuration
type GHTKConfig struct {
	BaseURL    string `json:"base_url"`
	Token      string `json:"token"`
	ShopID     string `json:"shop_id"`
	Timeout    int    `json:"timeout"` // in seconds
	IsTestMode bool   `json:"is_test_mode"`
}

// GHTKClient handles GHTK API interactions
type GHTKClient struct {
	config     GHTKConfig
	httpClient *http.Client
}

// NewGHTKClient creates a new GHTK client
func NewGHTKClient(config GHTKConfig) *GHTKClient {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GHTKClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GHTK API Request/Response Models

// CreateOrderRequest represents GHTK create order request
type CreateOrderRequest struct {
	Products []GHTKProduct `json:"products"`
	Order    GHTKOrder     `json:"order"`
}

type GHTKProduct struct {
	Name        string  `json:"name"`
	Weight      int     `json:"weight"` // in grams
	Quantity    int     `json:"quantity"`
	ProductCode string  `json:"product_code"`
	Price       float64 `json:"price"`
}

type GHTKOrder struct {
	ID             string   `json:"id"`              // Order ID from our system
	PickName       string   `json:"pick_name"`       // Tên người lấy hàng
	PickAddress    string   `json:"pick_address"`    // Địa chỉ lấy hàng
	PickProvince   string   `json:"pick_province"`   // Tỉnh/TP lấy hàng
	PickDistrict   string   `json:"pick_district"`   // Quận/Huyện lấy hàng
	PickWard       string   `json:"pick_ward"`       // Phường/Xã lấy hàng
	PickStreet     string   `json:"pick_street"`     // Đường lấy hàng
	PickTel        string   `json:"pick_tel"`        // SĐT người lấy hàng
	PickEmail      string   `json:"pick_email"`      // Email người lấy hàng
	Name           string   `json:"name"`            // Tên người nhận
	Address        string   `json:"address"`         // Địa chỉ nhận hàng
	Province       string   `json:"province"`        // Tỉnh/TP nhận hàng
	District       string   `json:"district"`        // Quận/Huyện nhận hàng
	Ward           string   `json:"ward"`            // Phường/Xã nhận hàng
	Street         string   `json:"street"`          // Đường nhận hàng
	Tel            string   `json:"tel"`             // SĐT người nhận
	Email          string   `json:"email"`           // Email người nhận
	Note           string   `json:"note"`            // Ghi chú
	Value          int      `json:"value"`           // Giá trị hàng hóa (VND)
	Transport      string   `json:"transport"`       // Phương thức vận chuyển
	PickOption     string   `json:"pick_option"`     // Tùy chọn lấy hàng
	DeliverOption  string   `json:"deliver_option"`  // Tùy chọn giao hàng
	PickSession    int      `json:"pick_session"`    // Ca lấy hàng
	DeliverSession int      `json:"deliver_session"` // Ca giao hàng
	LabelID        string   `json:"label_id"`        // ID nhãn đơn hàng
	PickMoney      int      `json:"pick_money"`      // Tiền thu hộ
	IsFreeship     int      `json:"is_freeship"`     // Miễn phí ship
	WeightOption   string   `json:"weight_option"`   // Tùy chọn cân nặng
	Tags           []string `json:"tags"`            // Nhãn đặc biệt
}

// CreateOrderResponse represents GHTK create order response
type CreateOrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Order   struct {
		LabelID              string        `json:"label_id"`
		PartnerID            string        `json:"partner_id"`
		Status               string        `json:"status"`
		Fee                  int           `json:"fee"`
		InsuranceFee         int           `json:"insurance_fee"`
		EstimatedPickTime    string        `json:"estimated_pick_time"`
		EstimatedDeliverTime string        `json:"estimated_deliver_time"`
		Products             []GHTKProduct `json:"products"`
		StatusID             int           `json:"status_id"`
		Created              string        `json:"created"`
		Updated              string        `json:"updated"`
	} `json:"order"`
}

// CalculateFeeRequest represents GHTK calculate fee request
type CalculateFeeRequest struct {
	PickProvince string `json:"pick_province"`
	PickDistrict string `json:"pick_district"`
	PickWard     string `json:"pick_ward"`
	Province     string `json:"province"`
	District     string `json:"district"`
	Ward         string `json:"ward"`
	Value        int    `json:"value"`
	Transport    string `json:"transport"`
	Weight       int    `json:"weight"`
}

// CalculateFeeResponse represents GHTK calculate fee response
type CalculateFeeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Fee     struct {
		Name         string `json:"name"`
		Fee          int    `json:"fee"`
		InsuranceFee int    `json:"insurance_fee"`
		TotalFee     int    `json:"total_fee"`
	} `json:"fee"`
}

// OrderStatusResponse represents GHTK order status response
type OrderStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Order   struct {
		LabelID     string         `json:"label_id"`
		PartnerID   string         `json:"partner_id"`
		Status      string         `json:"status"`
		StatusText  string         `json:"status_text"`
		Created     string         `json:"created"`
		Updated     string         `json:"updated"`
		PickDate    string         `json:"pick_date"`
		DeliverDate string         `json:"deliver_date"`
		Products    []GHTKProduct  `json:"products"`
		Timeline    []GHTKTimeline `json:"timeline"`
	} `json:"order"`
}

type GHTKTimeline struct {
	Status     string `json:"status"`
	StatusText string `json:"status_text"`
	Time       string `json:"time"`
	Location   string `json:"location"`
	Note       string `json:"note"`
}

// CancelOrderRequest represents GHTK cancel order request
type CancelOrderRequest struct {
	LabelID string `json:"label_id"`
	Note    string `json:"note"`
}

// CancelOrderResponse represents GHTK cancel order response
type CancelOrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WebhookData represents GHTK webhook data
type WebhookData struct {
	LabelID     string         `json:"label_id"`
	PartnerID   string         `json:"partner_id"`
	Status      string         `json:"status"`
	StatusText  string         `json:"status_text"`
	Created     string         `json:"created"`
	Updated     string         `json:"updated"`
	PickDate    string         `json:"pick_date"`
	DeliverDate string         `json:"deliver_date"`
	Products    []GHTKProduct  `json:"products"`
	Timeline    []GHTKTimeline `json:"timeline"`
}

// GHTK API Methods

// CreateOrder creates a new order in GHTK
func (c *GHTKClient) CreateOrder(req *CreateOrderRequest) (*CreateOrderResponse, error) {
	url := c.getAPIURL("/services/shipment/order")

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response CreateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

// CalculateFee calculates shipping fee
func (c *GHTKClient) CalculateFee(req *CalculateFeeRequest) (*CalculateFeeResponse, error) {
	url := c.getAPIURL("/services/shipment/fee")

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response CalculateFeeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

// GetOrderStatus gets order status from GHTK
func (c *GHTKClient) GetOrderStatus(labelID string) (*OrderStatusResponse, error) {
	url := c.getAPIURL(fmt.Sprintf("/services/shipment/v2/%s", labelID))

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response OrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

// CancelOrder cancels an order in GHTK
func (c *GHTKClient) CancelOrder(req *CancelOrderRequest) (*CancelOrderResponse, error) {
	url := c.getAPIURL("/services/shipment/cancel")

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response CancelOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

// PrintLabel prints shipping label
func (c *GHTKClient) PrintLabel(labelID string) ([]byte, error) {
	url := c.getAPIURL(fmt.Sprintf("/services/shipment/label/%s", labelID))

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	labelData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read label data: %v", err)
	}

	return labelData, nil
}

// Helper methods

func (c *GHTKClient) getAPIURL(endpoint string) string {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		if c.config.IsTestMode {
			baseURL = "https://dev.ghtk.vn"
		} else {
			baseURL = "https://services.ghtk.vn"
		}
	}
	return baseURL + endpoint
}

func (c *GHTKClient) makeRequest(method, url string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
	}

	// Add GHTK headers
	req.Header.Set("Token", c.config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// VerifyWebhookSignature verifies GHTK webhook signature
func (c *GHTKClient) VerifyWebhookSignature(signature string, body []byte) bool {
	// GHTK webhook verification logic would go here
	// For now, return true (implement proper verification)
	return true
}
