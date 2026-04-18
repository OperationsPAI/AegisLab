package httpx

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestRequestIDFromContextPrefersLocalValue(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "from-md"))
	ctx = WithRequestID(ctx, "from-context")

	if got := RequestIDFromContext(ctx); got != "from-context" {
		t.Fatalf("RequestIDFromContext() = %q, want %q", got, "from-context")
	}
}

func TestRequestIDFromContextFallsBackToIncomingMetadata(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "from-md"))

	if got := RequestIDFromContext(ctx); got != "from-md" {
		t.Fatalf("RequestIDFromContext() = %q, want %q", got, "from-md")
	}
}

func TestWithOutgoingRequestIDCopiesValueToMetadata(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-123")

	outgoing := WithOutgoingRequestID(ctx)
	md, ok := metadata.FromOutgoingContext(outgoing)
	if !ok {
		t.Fatal("expected outgoing metadata to exist")
	}

	values := md.Get("x-request-id")
	if len(values) != 1 || values[0] != "req-123" {
		t.Fatalf("outgoing request id = %v, want [req-123]", values)
	}
}

func TestRequestIDFromContextFallsBackToOutgoingMetadata(t *testing.T) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-request-id", "outgoing-md"))

	if got := RequestIDFromContext(ctx); got != "outgoing-md" {
		t.Fatalf("RequestIDFromContext() = %q, want %q", got, "outgoing-md")
	}
}
