package consumer

import (
	buildkitinfra "aegis/infra/buildkit"
	helminfra "aegis/infra/helm"
	k8sinfra "aegis/infra/k8s"
	redisinfra "aegis/infra/redis"

	"gorm.io/gorm"
)

type RuntimeDeps struct {
	DB                   *gorm.DB
	Monitor              NamespaceMonitor
	RestartRateLimiter   *TokenBucketRateLimiter
	BuildRateLimiter     *TokenBucketRateLimiter
	AlgorithmRateLimiter *TokenBucketRateLimiter
	RedisGateway         *redisinfra.Gateway
	K8sGateway           *k8sinfra.Gateway
	BuildKitGateway      *buildkitinfra.Gateway
	HelmGateway          *helminfra.Gateway
	FaultBatchManager    *FaultBatchManager
	ExecutionOwner       ExecutionOwner
	InjectionOwner       InjectionOwner
}
