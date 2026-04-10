package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"

	rk "github.com/reststore/restkit"
	rkhttp3 "github.com/reststore/restkit/extra/http3"
)

type PingResponse struct {
	Message string `json:"message"`
}

func main() {
	if err := generateCert("cert.pem", "key.pem"); err != nil {
		log.Fatal(err)
	}

	api := rk.NewApi()
	api.WithSwaggerUI()
	api.WithVersion("1.0.0")
	api.WithTitle("HTTP/3 Example API")

	// Define a simple ping endpoint to test the API.
	ping := rk.Get("/ping",
		func(ctx context.Context, _ rk.NoRequest) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		},
	)

	api.AddEndpoint(ping)

	// Add both HTTP/2 and HTTP/3 servers to OAS for swagger.
	api.WithServer("https://localhost:8080", "Local Dev Server (HTTP/2)", nil)
	api.WithServer("https://localhost:8081", "Local Dev Server (HTTP/3)", nil)

	log.Println("Starting HTTP/2 + HTTP/3 server on :8080 (TCP) and :8081 (UDP)...")
	log.Println("API Documentation: https://localhost:8080/swagger")

	// Serve the api using http2 and http3 with TLS certificates
	if err := rkhttp3.Serve(api, ":8080", ":8081", "cert.pem", "key.pem"); err != nil {
		log.Fatal(err)
	}
}

// generateCert creates self-signed TLS certificates
// for testing purposes if they don't already exist.
func generateCert(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil
		}
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           nil,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return nil
}
