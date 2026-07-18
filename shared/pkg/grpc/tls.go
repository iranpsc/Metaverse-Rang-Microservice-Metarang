package grpcutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TLSEnabled reports whether inter-service TLS/mTLS is enabled.
func TLSEnabled() bool {
	value := os.Getenv("GRPC_TLS_ENABLED")
	if value == "" {
		return false
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return enabled
}

// ServerCredentials returns transport credentials for gRPC servers.
// When GRPC_TLS_ENABLED is false, returns nil (plaintext).
func ServerCredentials() (credentials.TransportCredentials, error) {
	if !TLSEnabled() {
		return nil, nil
	}

	certFile := os.Getenv("GRPC_TLS_CERT_FILE")
	keyFile := os.Getenv("GRPC_TLS_KEY_FILE")
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("GRPC_TLS_ENABLED=true requires GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load server TLS key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if caFile := os.Getenv("GRPC_TLS_CA_FILE"); caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read GRPC_TLS_CA_FILE: %w", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse GRPC_TLS_CA_FILE")
		}
		tlsConfig.ClientCAs = certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return credentials.NewTLS(tlsConfig), nil
}

// ClientCredentials returns transport credentials for gRPC clients.
func ClientCredentials() (credentials.TransportCredentials, error) {
	if !TLSEnabled() {
		return insecure.NewCredentials(), nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if caFile := os.Getenv("GRPC_TLS_CA_FILE"); caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read GRPC_TLS_CA_FILE: %w", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse GRPC_TLS_CA_FILE")
		}
		tlsConfig.RootCAs = certPool
	}

	clientCertFile := os.Getenv("GRPC_TLS_CLIENT_CERT_FILE")
	clientKeyFile := os.Getenv("GRPC_TLS_CLIENT_KEY_FILE")
	if clientCertFile != "" && clientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client TLS key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return credentials.NewTLS(tlsConfig), nil
}
