package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestFileChangeNotification(t *testing.T) {
	go main()
	time.Sleep(1 * time.Second) // Give server time to start

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Because self-signed cert
			},
		},
	}

	req, err := http.NewRequest("GET", "https://localhost:8443/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Simulate file modification after short delay
	go func() {
		time.Sleep(1 * time.Second)
		os.WriteFile("watched.txt", []byte("changed"), 0644)
	}()

	buf := make([]byte, 4096)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("No data received")
	}
	t.Logf("Received: %s", string(buf[:n]))
}
