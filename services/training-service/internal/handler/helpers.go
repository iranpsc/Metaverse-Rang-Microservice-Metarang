package handler

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// IPAddressFromGRPCContext reads client IP from gRPC incoming metadata (gateway/proxy headers).
func IPAddressFromGRPCContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			return ips[0]
		}
		if ips := md.Get("x-real-ip"); len(ips) > 0 {
			return ips[0]
		}
	}
	return "unknown"
}
