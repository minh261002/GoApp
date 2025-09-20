package payment

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go_app/pkg/logger"
)

// PayOSConfig holds PayOS configuration
type PayOSConfig struct {
	ClientID    string
	APIKey      string
	ChecksumKey string
	BaseURL     string
}

// PayOSClient handles PayOS API interactions
type PayOSClient struct {
	config     PayOSConfig
	httpClient *http.Client
}

// PayOSItem represents an item in the payment
type PayOSItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"` // Price in VND
}

// PayOSPaymentData represents payment data for PayOS
type PayOSPaymentData struct {
	OrderCode   int         `json:"orderCode"`
	Amount      int         `json:"amount"` // Amount in VND
	Description string      `json:"description"`
	Items       []PayOSItem `json:"items"`
	ReturnURL   string      `json:"returnUrl"`
	CancelURL   string      `json:"cancelUrl"`
	ExpiredAt   *int64      `json:"expiredAt,omitempty"` // Unix timestamp
	Signature   string      `json:"signature,omitempty"`
}

// PayOSCreatePaymentResponse represents the response from PayOS create payment API
type PayOSCreatePaymentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Bin           string `json:"bin"`
		CheckoutURL   string `json:"checkoutUrl"`
		AccountNumber string `json:"accountNumber"`
		AccountName   string `json:"accountName"`
		Amount        int    `json:"amount"`
		Description   string `json:"description"`
		OrderCode     int    `json:"orderCode"`
		QRCode        string `json:"qrCode"`
	} `json:"data"`
}

// PayOSPaymentInfo represents payment information from PayOS
type PayOSPaymentInfo struct {
	OrderCode            int    `json:"orderCode"`
	Amount               int    `json:"amount"`
	Description          string `json:"description"`
	AccountNumber        string `json:"accountNumber"`
	Reference            string `json:"reference"`
	TransactionID        string `json:"transactionId"`
	Code                 string `json:"code"`
	Desc                 string `json:"desc"`
	CounterAccountBankId string `json:"counterAccountBankId"`
	VirtualAccountName   string `json:"virtualAccountName"`
	VirtualAccountNumber string `json:"virtualAccountNumber"`
}

// PayOSWebhookData represents webhook data from PayOS
type PayOSWebhookData struct {
	Code      int              `json:"code"`
	Message   string           `json:"message"`
	Data      PayOSPaymentInfo `json:"data"`
	Signature string           `json:"signature"`
}

// NewPayOSClient creates a new PayOS client
func NewPayOSClient(config PayOSConfig) *PayOSClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api-merchant.payos.vn"
	}

	return &PayOSClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePaymentLink creates a payment link via PayOS
func (c *PayOSClient) CreatePaymentLink(paymentData PayOSPaymentData) (*PayOSCreatePaymentResponse, error) {
	// Generate signature
	signature, err := c.generateSignature(paymentData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %v", err)
	}
	paymentData.Signature = signature

	// Convert to JSON
	jsonData, err := json.Marshal(paymentData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment data: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.config.BaseURL+"/v2/payment-requests", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-client-id", c.config.ClientID)
	req.Header.Set("x-api-key", c.config.APIKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var response PayOSCreatePaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("PayOS API error: %s", response.Message)
	}

	logger.Infof("PayOS payment link created successfully: OrderCode=%d, CheckoutURL=%s",
		response.Data.OrderCode, response.Data.CheckoutURL)

	return &response, nil
}

// GetPaymentInfo gets payment information from PayOS
func (c *PayOSClient) GetPaymentInfo(orderCode int) (*PayOSPaymentInfo, error) {
	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/payment-requests/%d", c.config.BaseURL, orderCode), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("x-client-id", c.config.ClientID)
	req.Header.Set("x-api-key", c.config.APIKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var response struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    PayOSPaymentInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("PayOS API error: %s", response.Message)
	}

	return &response.Data, nil
}

// CancelPayment cancels a payment
func (c *PayOSClient) CancelPayment(orderCode int, cancellationReason string) error {
	// Prepare cancellation data
	cancelData := map[string]interface{}{
		"cancellationReason": cancellationReason,
	}

	jsonData, err := json.Marshal(cancelData)
	if err != nil {
		return fmt.Errorf("failed to marshal cancellation data: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/payment-requests/%d/cancel", c.config.BaseURL, orderCode), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-client-id", c.config.ClientID)
	req.Header.Set("x-api-key", c.config.APIKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		return fmt.Errorf("PayOS API error: %s", response.Message)
	}

	logger.Infof("PayOS payment cancelled successfully: OrderCode=%d", orderCode)
	return nil
}

// VerifyWebhookSignature verifies webhook signature from PayOS
func (c *PayOSClient) VerifyWebhookSignature(signature string, data []byte) bool {
	expectedSignature := c.generateWebhookSignature(data)
	return signature == expectedSignature
}

// generateSignature generates signature for PayOS API
func (c *PayOSClient) generateSignature(paymentData PayOSPaymentData) (string, error) {
	// Create data string for signature
	dataStr := fmt.Sprintf("amount=%d&cancelUrl=%s&description=%s&orderCode=%d&returnUrl=%s",
		paymentData.Amount,
		paymentData.CancelURL,
		paymentData.Description,
		paymentData.OrderCode,
		paymentData.ReturnURL,
	)

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(c.config.ChecksumKey))
	h.Write([]byte(dataStr))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, nil
}

// generateWebhookSignature generates signature for webhook verification
func (c *PayOSClient) generateWebhookSignature(data []byte) string {
	h := hmac.New(sha256.New, []byte(c.config.ChecksumKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// ConvertVNDToInt converts VND amount to int (multiply by 100 for PayOS)
func ConvertVNDToInt(amount float64) int {
	return int(amount * 100)
}

// ConvertIntToVND converts int amount to VND (divide by 100)
func ConvertIntToVND(amount int) float64 {
	return float64(amount) / 100
}

// GenerateOrderCode generates a unique order code for PayOS
func GenerateOrderCode() int {
	return int(time.Now().Unix())
}

// IsValidPayOSResponse checks if PayOS response is valid
func IsValidPayOSResponse(response *PayOSCreatePaymentResponse) bool {
	return response != nil && response.Code == 0 && response.Data.CheckoutURL != ""
}
