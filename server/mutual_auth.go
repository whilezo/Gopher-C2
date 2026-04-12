package server

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadTLSServerCreds() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("client.crt")
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return credentials.NewTLS(tlsConfig), nil
}
