package main

// Demo WebTransport server using QUIC and HTTP/3
// This server listens on port 4433 and handles WebTransport sessions at the /webtransport endpoint.
// It uses a self-signed certificate for TLS, so you need to add the certificate to your browser's trusted certificates.
// With Brave, I use the following command to start the browser with the self-signed certificate trusted:
// .\brave.exe --origin-to-force-quic-on=localhost:4433 --ignore-certificate-errors-spki-list=ibdElbpy/Cl9ZssVrMvsLeXPIGPBHTv/N6KXObqeuKg=

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	// Load TLS certificate and key
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	// Configure the WebTransport server
	wtServer := &webtransport.Server{
		H3: http3.Server{
			Addr: ":4433",
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				NextProtos:   []string{"h3"},
			},
		},
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for demo!
			return true
		},
	}

	// Handle WebTransport sessions at the /webtransport endpoint
	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		session, err := wtServer.Upgrade(w, r)
		if err != nil {
			log.Printf("failed to upgrade: %v", err)
			http.Error(w, "failed to upgrade", http.StatusInternalServerError)
			return
		}
		go handleSession(session)
	})

	// Start the server
	log.Println("Starting WebTransport server on :4433")
	if err := wtServer.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func handleSession(session *webtransport.Session) {
	defer session.CloseWithError(0, "session closed")

	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Printf("failed to accept stream: %v", err)
			return
		}
		go handleStream(stream)
	}
}

func handleStream(stream webtransport.Stream) {
	defer stream.Close()

	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Printf("failed to read from stream: %v", err)
			return
		}
		log.Printf("Received: %s", string(buf[:n]))

		_, err = stream.Write(buf[:n])
		if err != nil {
			log.Printf("failed to write to stream: %v", err)
			return
		}
	}
}
