package apns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestAPNsSender_Send2(t *testing.T) {

	// Load configuration
	var config Config
	if err := loadConfig("../../config.json", &config); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Example Alert struct
	alert := Alert{
		Device:   "010fbd4480b8f3915b66035fa0ffce31e30b73fd802a61d55b917ac997ab4ce7", // Replace with a valid device token
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

func TestAPNsSender_Send(t *testing.T) {
	// Create a test config
	config := &Config{
		TeamID:     "TEST_TEAM_ID",
		KeyID:      "TEST_KEY_ID",
		P8KeyPath:  "test_key.p8", // We will create a dummy key file
		BundleID:   "com.example.test",
		Production: false,
	}

	// Create a dummy p8 key file for testing
	dummyKey := []byte(`-----BEGIN PRIVATE KEY-----
	DUMMY_PRIVATE_KEY
	-----END PRIVATE KEY-----`)
	os.WriteFile("test_key.p8", dummyKey, 0644)
	defer os.Remove("test_key.p8") // Clean up after test
	// Load configuration
	// var config Config
	if err := loadConfig("../../config.json", &config); err != nil {
		log.Fatalf("Config error: %v", err)
	}
	// Add debug output
	fmt.Printf("Loaded config: %+v\n", config)
	sender, err := NewAPNsSender(config)
	if err != nil {
		t.Fatalf("Failed to create APNs sender: %v", err)
	}

	testAlert := &Alert{
		Device:   "010fbd4480b8f3915b66035fa0ffce31e30b73fd802a61d55b917ac997ab4ce7",
		Title:    "Test Title",
		Subtitle: "Test Subtitle",
		Body:     "Test Body",
		Action:   "test_action",
		Param:    "test_param",
		URL:      "test_url", // URL field is not used in the payload structure as per current code, using Action for URL in payload
	}

	// Mock HTTP Client
	mockHTTPClient := &http.Client{
		Transport: &mockTransport{
			Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"success":"OK"}`)),
			},
		},
	}
	sender.client = mockHTTPClient // Inject mock client

	err = sender.Send(testAlert)
	if err != nil {
		t.Errorf("Send() error = %v, wantErr nil", err)
	}

	// Test for error response
	mockErrorHTTPClient := &http.Client{
		Transport: &mockTransport{
			Response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"reason":"BadDeviceToken"}`)),
			},
		},
	}
	sender.client = mockErrorHTTPClient // Inject mock error client

	err = sender.Send(testAlert)
	if err == nil {
		t.Errorf("Send() error = nil, wantErr not nil for error response")
	} else {
		if err.Error() != "APNs request failed with status 400: {\"reason\":\"BadDeviceToken\"}" {
			t.Errorf("Send() error message mismatch: got %v, want %v", err.Error(), "APNs request failed with status 400: {\"reason\":\"BadDeviceToken\"}")
		}
	}
}

// mockTransport for testing HTTP requests
type mockTransport struct {
	Response *http.Response
	Error    error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Assert request details if needed, e.g., headers, URL, payload
	if req.Header.Get("apns-topic") != "com.example.test" {
		return nil, fmt.Errorf("incorrect apns-topic header")
	}

	var payload struct {
		APS struct {
			Alert struct {
				Title    string `json:"title,omitempty"`
				Subtitle string `json:"subtitle,omitempty"`
				Body     string `json:"body,omitempty"`
			} `json:"alert,omitempty"`
			URL string `json:"url,omitempty"`
		} `json:"aps"`
		Param string `json:"param,omitempty"`
	}
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}
	if payload.APS.Alert.Title != "Test Title" {
		return nil, fmt.Errorf("incorrect payload title")
	}
	if payload.Param != "test_param" {
		return nil, fmt.Errorf("incorrect payload param")
	}
	if payload.APS.URL != "test_action" { // Asserting Action as URL
		return nil, fmt.Errorf("incorrect payload URL (action)")
	}

	return m.Response, m.Error
}
