package app

import (
	"testing"

	"github.com/spf13/viper"
)

func TestMissingRequiredTargets(t *testing.T) {
	primaryKey := "clients.runtime.target"
	legacyKey := "runtime_worker.grpc.target"

	originalPrimary := viper.Get(primaryKey)
	originalLegacy := viper.Get(legacyKey)
	t.Cleanup(func() {
		viper.Set(primaryKey, originalPrimary)
		viper.Set(legacyKey, originalLegacy)
	})

	viper.Set(primaryKey, "")
	viper.Set(legacyKey, "")

	missing := missingRequiredTargets(RequiredConfigTarget{
		Name:       "runtime-worker-service",
		PrimaryKey: primaryKey,
		LegacyKey:  legacyKey,
	})
	if len(missing) != 1 {
		t.Fatalf("expected 1 missing target, got %d: %v", len(missing), missing)
	}

	viper.Set(primaryKey, "127.0.0.1:9094")
	missing = missingRequiredTargets(RequiredConfigTarget{
		Name:       "runtime-worker-service",
		PrimaryKey: primaryKey,
		LegacyKey:  legacyKey,
	})
	if len(missing) != 0 {
		t.Fatalf("expected no missing target when primary key is set, got %v", missing)
	}

	viper.Set(primaryKey, "")
	viper.Set(legacyKey, "127.0.0.1:9094")
	missing = missingRequiredTargets(RequiredConfigTarget{
		Name:       "runtime-worker-service",
		PrimaryKey: primaryKey,
		LegacyKey:  legacyKey,
	})
	if len(missing) != 0 {
		t.Fatalf("expected no missing target when legacy key is set, got %v", missing)
	}
}
