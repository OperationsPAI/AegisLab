package httpx

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	RequestIDHeader      = "X-Request-Id"
	requestIDMetadataKey = "x-request-id"
)

type requestIDContextKey struct{}

func NewRequestID() string {
	return uuid.NewString()
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ctx
	}

	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if requestID, ok := ctx.Value(requestIDContextKey{}).(string); ok && strings.TrimSpace(requestID) != "" {
		return strings.TrimSpace(requestID)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID := firstMetadataValue(md, requestIDMetadataKey); requestID != "" {
			return requestID
		}
	}

	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if requestID := firstMetadataValue(md, requestIDMetadataKey); requestID != "" {
			return requestID
		}
	}

	return ""
}

func WithOutgoingRequestID(ctx context.Context) context.Context {
	requestID := RequestIDFromContext(ctx)
	if requestID == "" {
		return ctx
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set(requestIDMetadataKey, requestID)
	return metadata.NewOutgoingContext(WithRequestID(ctx, requestID), md)
}

func UnaryClientRequestIDInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(WithOutgoingRequestID(ctx), method, req, reply, cc, opts...)
	}
}

func UnaryServerRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		requestID := RequestIDFromContext(ctx)
		if requestID == "" {
			requestID = NewRequestID()
		}

		ctx = WithRequestID(ctx, requestID)
		if err := grpc.SetHeader(ctx, metadata.Pairs(requestIDMetadataKey, requestID)); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func firstMetadataValue(md metadata.MD, key string) string {
	if md == nil {
		return ""
	}

	values := md.Get(key)
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
