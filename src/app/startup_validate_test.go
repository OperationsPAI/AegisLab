package app

import (
	"testing"

	"go.uber.org/fx"
)

func TestProducerOptionsValidate(t *testing.T) {
	if err := fx.ValidateApp(ProducerOptions("..", "0")); err != nil {
		t.Fatalf("producer fx graph validation failed: %v", err)
	}
}

func TestConsumerOptionsValidate(t *testing.T) {
	if err := fx.ValidateApp(ConsumerOptions("..")); err != nil {
		t.Fatalf("consumer fx graph validation failed: %v", err)
	}
}

func TestBothOptionsValidate(t *testing.T) {
	if err := fx.ValidateApp(BothOptions("..", "0")); err != nil {
		t.Fatalf("both fx graph validation failed: %v", err)
	}
}
