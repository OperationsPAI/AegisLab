package app

import (
	chaosinfra "aegis/infra/chaos"
	k8sinfra "aegis/infra/k8s"
	httpinterface "aegis/interface/http"

	"go.uber.org/fx"
)

// ProducerCompatibilityOptions captures the standalone producer/api-gateway
// HTTP stack, including the HTTP-side K8s/chaos infra.
func ProducerCompatibilityOptions(port string) fx.Option {
	return fx.Options(
		chaosinfra.Module,
		k8sinfra.Module,
		ProducerHTTPEntryOptions(port),
	)
}

// ProducerHTTPEntryOptions captures the compatibility producer HTTP surface
// shared by producer, both, and api-gateway entrypoints.
func ProducerHTTPEntryOptions(port string) fx.Option {
	return fx.Options(
		fx.Provide(newProducerInitializer),
		fx.Invoke(registerProducerInitialization),
		ProducerHTTPModules(),
		fx.Supply(httpinterface.ServerConfig{Addr: normalizeAddr(port)}),
		httpinterface.Module,
	)
}

// CompatibilityRuntimeOptions centralizes the legacy runtime stack that still
// needs local execution/injection owners for producer/consumer/both entrypoints.
func CompatibilityRuntimeOptions() fx.Option {
	return fx.Options(
		RuntimeWorkerStackOptions(),
		ExecutionInjectionOwnerModules(),
	)
}

// BothCompatibilityOptions captures the legacy combined producer+consumer
// runtime surface in one place.
func BothCompatibilityOptions(port string) fx.Option {
	return fx.Options(
		CompatibilityRuntimeOptions(),
		ProducerHTTPEntryOptions(port),
	)
}
