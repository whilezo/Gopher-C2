package server

import (
	"context"
	"net"

	"google.golang.org/grpc/peer"
)

func getClientIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}

	// p.Addr is a net.Addr interface.
	// We usually expect a *net.TCPAddr for gRPC over TCP.
	if tcpAddr, ok := p.Addr.(*net.TCPAddr); ok {
		return tcpAddr.IP.String()
	}

	return p.Addr.String()
}
