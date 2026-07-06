package handler

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestIPAddressFromGRPCContext_ForwardedFor(t *testing.T) {
	md := metadata.Pairs("x-forwarded-for", "203.0.113.1")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	if got := IPAddressFromGRPCContext(ctx); got != "203.0.113.1" {
		t.Fatal(got)
	}
}

func TestIPAddressFromGRPCContext_RealIP(t *testing.T) {
	md := metadata.Pairs("x-real-ip", "198.51.100.2")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	if got := IPAddressFromGRPCContext(ctx); got != "198.51.100.2" {
		t.Fatal(got)
	}
}

func TestIPAddressFromGRPCContext_Unknown(t *testing.T) {
	if got := IPAddressFromGRPCContext(context.Background()); got != "unknown" {
		t.Fatal(got)
	}
}
