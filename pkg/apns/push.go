package apns

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// APNs hosts
const (
	APNsProduction    = "api.push.apple.com"
	APNsDevelopment   = "api.sandbox.push.apple.com"
	APNsTokenEndpoint = "/3/device/"
)

// Config holds the APNs configuration
type Config struct {
	TeamID     string `json:"team_id"`
	KeyID      string `json:"key_id"`
	P8KeyPath  string `json:"p8"`
	BundleID   string `json:"bundle_id"`
	Production bool   `json:"production"`
}

// Alert holds the notification payload
type Alert struct {
	Device   string `json:"device"`
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Body     string `json:"body,omitempty"`
	Action   string `json:"action,omitempty"`
	Param    string `json:"param,omitempty"`
	URL      string `json:"url,omitempty"`
}

// PushSender interface defines the notification sender contract
type PushSender interface {
	Send(alert *Alert) error
}

// APNsSender implements PushSender for Apple Push Notification service
type APNsSender struct {
	config     *Config
	client     *http.Client
	privateKey *ecdsa.PrivateKey
}

// NewAPNsSender creates a new APNs sender instance
func NewAPNsSender(config *Config) (*APNsSender, error) {
	// Load private key
	p8Data, err := os.ReadFile(config.P8KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read p8 file: %v", err)
	}

	block, _ := pem.Decode(p8Data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block from p8 file")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	privateKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected ECDSA private key, got %T", key)
	}

	return &APNsSender{
		config:     config,
		client:     &http.Client{},
		privateKey: privateKey,
	}, nil
}

// Send sends an APNs push notification
func (a *APNsSender) Send(alert *Alert) error {
	// Generate JWT token
	token, err := a.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %v", err)
	}

	// Prepare request
	host := APNsDevelopment
	if a.config.Production {
		host = APNsProduction
	}

	url := fmt.Sprintf("https://%s%s%s", host, APNsTokenEndpoint, alert.Device)

	payload := struct {
		APS struct {
			Alert struct {
				Title    string `json:"title,omitempty"`
				Subtitle string `json:"subtitle,omitempty"`
				Body     string `json:"body,omitempty"`
			} `json:"alert,omitempty"`
			URL string `json:"url,omitempty"`
		} `json:"aps"`
		Param string `json:"param,omitempty"`
	}{
		APS: struct {
			Alert struct {
				Title    string `json:"title,omitempty"`
				Subtitle string `json:"subtitle,omitempty"`
				Body     string `json:"body,omitempty"`
			} `json:"alert,omitempty"`
			URL string `json:"url,omitempty"`
		}{
			Alert: struct {
				Title    string `json:"title,omitempty"`
				Subtitle string `json:"subtitle,omitempty"`
				Body     string `json:"body,omitempty"`
			}{
				Title:    alert.Title,
				Subtitle: alert.Subtitle,
				Body:     alert.Body,
			},
			URL: alert.Action, // Using Action as URL as per prompt
		},
		Param: alert.Param,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("apns-topic", a.config.BundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")
	req.Header.Set("authorization", fmt.Sprintf("bearer %s", token))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Handle response
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[APNs] Status Code: %d\nResponse: %s", resp.StatusCode, string(body))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("APNs request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateToken creates the JWT authentication token for APNs
func (a *APNsSender) generateToken() (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    a.config.TeamID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = a.config.KeyID

	return token.SignedString(a.privateKey)
}

// loadConfig loads configuration from JSON file
func loadConfig(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return nil
}

// Push function sends an APNs notification in a goroutine
func Push(cfg Config, alert Alert) {
	go func() {
		sender, err := NewAPNsSender(&cfg)
		if err != nil {
			log.Printf("Failed to create APNs sender: %v", err)
			return
		}

		if err := sender.Send(&alert); err != nil {
			log.Printf("Failed to send notification: %v", err)
			return
		}

		log.Println("Notification sent successfully")
	}()
}

func main() {
	// Load configuration
	var config Config
	if err := loadConfig("config.json", &config); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Example Alert struct
	alert := Alert{
		Device:   "YOUR_DEVICE_TOKEN", // Replace with a valid device token
		Title:    "Hello",
		Subtitle: "From APNs Module",
		Body:     "This is a test notification sent using the updated main function.",
		Action:   "https://example.com/deep_link", // Example URL/Action
		Param:    "custom_param_value",            // Example custom parameter
	}

	Push(config, alert)

	fmt.Println("Push notification dispatched in goroutine...")
	// Keep main function running for a while to allow goroutine to finish (for demonstration purposes)
	time.Sleep(2 * time.Second)
}
