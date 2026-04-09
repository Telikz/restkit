package http3

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"github.com/quic-go/quic-go/http3"
	rk "github.com/reststore/restkit"
)

func Serve(api *rk.Api, tcpAddr, udpAddr, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("load TLS certificates: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2", "h3"},
	}

	mux := api.Mux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Go(func() {
		h3Server := &http3.Server{
			Addr:      udpAddr,
			Handler:   mux,
			TLSConfig: tlsConfig,
		}
		if err := h3Server.ListenAndServe(); err != nil {
			errChan <- fmt.Errorf("http3: %w", err)
			cancel()
		}
	})

	wg.Go(func() {
		h2Server := &http.Server{
			Addr:      tcpAddr,
			Handler:   mux,
			TLSConfig: tlsConfig,
		}
		if err := h2Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http2: %w", err)
			cancel()
		}
	})

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		wg.Wait()
		return nil
	}
}
