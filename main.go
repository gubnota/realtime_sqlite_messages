package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	mu          sync.Mutex
	subscribers = make([]chan string, 0)
)

func main() {
	// Watch the file
	go watchFile("watched.txt")

	// Setup HTTP/3 server
	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", subscribeHandler)

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	log.Println("HTTP/3 server starting on https://localhost:8443")
	err := server.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	notify := make(chan string)
	mu.Lock()
	subscribers = append(subscribers, notify)
	mu.Unlock()

	defer func() {
		mu.Lock()
		for i, c := range subscribers {
			if c == notify {
				// remove from list
				subscribers = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
		mu.Unlock()
		close(notify)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	ctx := r.Context()

	for {
		select {
		case msg := <-notify:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

func watchFile(filename string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Read the file content
				// This is a placeholder; replace with actual file reading logic
				// For example, you can use ioutil.ReadFile or os.Open
				// and io.Copy to read the file content into a buffer
				content, err := os.ReadFile(event.Name)
				if err != nil {
					log.Println("Error reading file:", err)
				}
				// Convert byte slice to string
				fileContent := string(content)
				log.Println("", fileContent)
				notifySubscribers(fmt.Sprintf("File %s changed", fileContent))
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func notifySubscribers(msg string) {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range subscribers {
		select {
		case c <- msg:
		default:
			// non-blocking
		}
	}
}

// package main

// import (
// 	"crypto/ecdsa"
// 	"crypto/elliptic"
// 	"crypto/rand"
// 	"crypto/tls"
// 	"crypto/x509"
// 	"crypto/x509/pkix"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"math/big"
// 	"net/http"
// 	"time"

// 	"github.com/fsnotify/fsnotify"
// 	"github.com/quic-go/quic-go/http3"
// )

// const (
// 	port = ":8080"
// 	// port        = ":443"
// 	// certFile    = "cert.pem"
// 	// keyFile     = "key.pem"
// 	apiKey         = "secret123"
// 	targetFile     = "watched.txt"
// 	requiredHeader = "X-API-Key"
// )

// // func main() {
// // http.HandleFunc("/watch", watchHandler)

// // log.Printf("Starting HTTP/3 server on %s", port)
// // err := http3.ListenAndServeQUIC(port, certFile, keyFile, nil)
// //
// //	if err != nil {
// //		log.Fatalf("Failed to start server: %v", err)
// //	}
// //
// // }
// func main() {
// 	// Generate self-signed certificate
// 	cert, privKey, err := generateSelfSignedCert()
// 	if err != nil {
// 		log.Fatalf("Failed to generate certificate: %v", err)
// 	}

// 	// Configure TLS with self-signed cert
// 	tlsConfig := &tls.Config{
// 		Certificates: []tls.Certificate{{
// 			Certificate: [][]byte{cert.Raw},
// 			PrivateKey:  privKey,
// 			Leaf:        cert,
// 		}},
// 		NextProtos: []string{"h3"},
// 	}

// 	// Create HTTP/3 server
// 	server := &http3.Server{
// 		Addr:      port,
// 		Handler:   http.DefaultServeMux,
// 		TLSConfig: tlsConfig,
// 	}

// 	http.HandleFunc("/watch", watchHandler)

// 	log.Printf("Starting HTTP/3 server with self-signed cert on %s", port)
// 	if err := server.ListenAndServe(); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }

// func generateSelfSignedCert() (*x509.Certificate, *ecdsa.PrivateKey, error) {
// 	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	template := x509.Certificate{
// 		SerialNumber: serialNumber,
// 		Subject: pkix.Name{
// 			Organization: []string{"Test Certificate"},
// 		},
// 		NotBefore:             time.Now(),
// 		NotAfter:              time.Now().Add(24 * time.Hour),
// 		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 		DNSNames:              []string{"localhost"},
// 	}

// 	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	cert, err := x509.ParseCertificate(certBytes)
// 	return cert, privKey, err
// }

// func watchHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Header.Get(requiredHeader) != apiKey {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")

// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// 	defer watcher.Close()

// 	if err := watcher.Add(targetFile); err != nil {
// 		http.Error(w, "File not found", http.StatusNotFound)
// 		return
// 	}

// 	// Send initial content
// 	if err := sendFileContents(w, flusher); err != nil {
// 		return
// 	}

// 	for {
// 		select {
// 		case event, ok := <-watcher.Events:
// 			if !ok {
// 				return
// 			}
// 			if event.Op&fsnotify.Write == fsnotify.Write {
// 				if err := sendFileContents(w, flusher); err != nil {
// 					return
// 				}
// 			}
// 		case err, ok := <-watcher.Errors:
// 			if !ok {
// 				return
// 			}
// 			log.Printf("Watcher error: %v", err)
// 		case <-r.Context().Done():
// 			log.Println("Client disconnected")
// 			return
// 		}
// 	}
// }

// func sendFileContents(w http.ResponseWriter, flusher http.Flusher) error {
// 	content, err := ioutil.ReadFile(targetFile)
// 	if err != nil {
// 		log.Printf("Error reading file: %v", err)
// 		http.Error(w, "Error reading file", http.StatusInternalServerError)
// 		return err
// 	}

// 	fmt.Fprintf(w, "data: %s\n\n", content)
// 	flusher.Flush()
// 	return nil
// }
