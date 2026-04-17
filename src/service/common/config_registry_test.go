package common

import (
	"testing"

	"aegis/consts"
)

func TestRegisterGlobalHandlersIsIdempotent(t *testing.T) {
	resetConfigRegistryForTest()
	t.Cleanup(resetConfigRegistryForTest)

	RegisterGlobalHandlers(nil)
	RegisterGlobalHandlers(nil)

	scope := consts.ConfigScopeGlobal
	keys := ListRegisteredConfigKeys(&scope)
	if len(keys) != 1 {
		t.Fatalf("expected 1 global handler, got %d: %v", len(keys), keys)
	}
	if keys[0] != "algo" {
		t.Fatalf("expected algo handler to be registered, got %v", keys)
	}
}
