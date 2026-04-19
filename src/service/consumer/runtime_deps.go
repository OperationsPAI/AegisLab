package consumer

import (
	buildkit "aegis/infra/buildkit"
	helm "aegis/infra/helm"
	k8s "aegis/infra/k8s"
	redis "aegis/infra/redis"

	"gorm.io/gorm"
)

type RuntimeDeps struct {
	DB                   *gorm.DB
	Monitor              NamespaceMonitor
	RestartRateLimiter   *TokenBucketRateLimiter
	BuildRateLimiter     *TokenBucketRateLimiter
	AlgorithmRateLimiter *TokenBucketRateLimiter
	RedisGateway         *redis.Gateway
	K8sGateway           *k8s.Gateway
	BuildKitGateway      *buildkit.Gateway
	HelmGateway          *helm.Gateway
	FaultBatchManager    *FaultBatchManager
	ExecutionOwner       ExecutionOwner
	InjectionOwner       InjectionOwner
}
