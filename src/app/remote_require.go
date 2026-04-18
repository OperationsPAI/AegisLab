package app

import (
	"context"
	"fmt"
	"strings"

	"aegis/config"

	"go.uber.org/fx"
)

type RequiredConfigTarget struct {
	Name       string
	PrimaryKey string
	LegacyKey  string
}

func RequireConfiguredTargets(component string, targets ...RequiredConfigTarget) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				missing := missingRequiredTargets(targets...)
				if len(missing) == 0 {
					return nil
				}
				return fmt.Errorf("%s requires configured internal client targets: %s", component, strings.Join(missing, ", "))
			},
		})
	})
}

func missingRequiredTargets(targets ...RequiredConfigTarget) []string {
	missing := make([]string, 0)
	for _, target := range targets {
		if target.PrimaryKey == "" {
			continue
		}

		primaryValue := strings.TrimSpace(config.GetString(target.PrimaryKey))
		legacyValue := strings.TrimSpace(config.GetString(target.LegacyKey))
		if primaryValue != "" || legacyValue != "" {
			continue
		}

		label := target.Name
		if label == "" {
			label = target.PrimaryKey
		}
		if target.LegacyKey != "" {
			label = fmt.Sprintf("%s (%s or %s)", label, target.PrimaryKey, target.LegacyKey)
		} else {
			label = fmt.Sprintf("%s (%s)", label, target.PrimaryKey)
		}
		missing = append(missing, label)
	}
	return missing
}
