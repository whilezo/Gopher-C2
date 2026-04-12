package client

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadTLSClientCreds() (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("server.crt")
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      pool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
