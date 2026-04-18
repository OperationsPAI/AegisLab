package common

import (
	"encoding/json"
	"fmt"
	"sync"

	"aegis/database"
	"aegis/repository"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// DBMetadataStore implements chaos.MetadataStore by reading from MySQL with in-memory caching.
type DBMetadataStore struct {
	cache sync.Map // key: "system:type:service" -> cached data
}

// NewDBMetadataStore creates a new DBMetadataStore instance.
func NewDBMetadataStore() *DBMetadataStore {
	return &DBMetadataStore{}
}

func (s *DBMetadataStore) cacheKey(system, metaType, service string) string {
	return system + ":" + metaType + ":" + service
}

func (s *DBMetadataStore) GetServiceEndpoints(system, serviceName string) ([]chaos.ServiceEndpointData, error) {
	key := s.cacheKey(system, "service_endpoint", serviceName)
	if cached, ok := s.cache.Load(key); ok {
		return cached.([]chaos.ServiceEndpointData), nil
	}

	meta, err := repository.GetSystemMetadata(database.DB, system, "service_endpoint", serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service endpoints: %w", err)
	}
	if meta == nil {
		return nil, nil
	}

	var result []chaos.ServiceEndpointData
	if err := json.Unmarshal([]byte(meta.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse service endpoint data: %w", err)
	}

	s.cache.Store(key, result)
	return result, nil
}

func (s *DBMetadataStore) GetAllServiceNames(system string) ([]string, error) {
	key := s.cacheKey(system, "_all_services", "")
	if cached, ok := s.cache.Load(key); ok {
		return cached.([]string), nil
	}

	names, err := repository.ListServiceNames(database.DB, system, "service_endpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %w", err)
	}

	s.cache.Store(key, names)
	return names, nil
}

func (s *DBMetadataStore) GetJavaClassMethods(system, serviceName string) ([]chaos.JavaClassMethodData, error) {
	key := s.cacheKey(system, "java_class_method", serviceName)
	if cached, ok := s.cache.Load(key); ok {
		return cached.([]chaos.JavaClassMethodData), nil
	}

	meta, err := repository.GetSystemMetadata(database.DB, system, "java_class_method", serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get java class methods: %w", err)
	}
	if meta == nil {
		return nil, nil
	}

	var result []chaos.JavaClassMethodData
	if err := json.Unmarshal([]byte(meta.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse java class method data: %w", err)
	}

	s.cache.Store(key, result)
	return result, nil
}

func (s *DBMetadataStore) GetDatabaseOperations(system, serviceName string) ([]chaos.DatabaseOperationData, error) {
	key := s.cacheKey(system, "database_operation", serviceName)
	if cached, ok := s.cache.Load(key); ok {
		return cached.([]chaos.DatabaseOperationData), nil
	}

	meta, err := repository.GetSystemMetadata(database.DB, system, "database_operation", serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database operations: %w", err)
	}
	if meta == nil {
		return nil, nil
	}

	var result []chaos.DatabaseOperationData
	if err := json.Unmarshal([]byte(meta.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse database operation data: %w", err)
	}

	s.cache.Store(key, result)
	return result, nil
}

func (s *DBMetadataStore) GetGRPCOperations(system, serviceName string) ([]chaos.GRPCOperationData, error) {
	key := s.cacheKey(system, "grpc_operation", serviceName)
	if cached, ok := s.cache.Load(key); ok {
		return cached.([]chaos.GRPCOperationData), nil
	}

	meta, err := repository.GetSystemMetadata(database.DB, system, "grpc_operation", serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC operations: %w", err)
	}
	if meta == nil {
		return nil, nil
	}

	var result []chaos.GRPCOperationData
	if err := json.Unmarshal([]byte(meta.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse gRPC operation data: %w", err)
	}

	s.cache.Store(key, result)
	return result, nil
}

func (s *DBMetadataStore) GetNetworkPairs(system string) ([]chaos.NetworkPairData, error) {
	key := s.cacheKey(system, "network_dependency", "_all")
	if cached, ok := s.cache.Load(key); ok {
		return cached.([]chaos.NetworkPairData), nil
	}

	metas, err := repository.ListSystemMetadata(database.DB, system, "network_dependency")
	if err != nil {
		return nil, fmt.Errorf("failed to get network pairs: %w", err)
	}

	var result []chaos.NetworkPairData
	for _, meta := range metas {
		var pairs []chaos.NetworkPairData
		if err := json.Unmarshal([]byte(meta.Data), &pairs); err != nil {
			return nil, fmt.Errorf("failed to parse network pair data: %w", err)
		}
		result = append(result, pairs...)
	}

	s.cache.Store(key, result)
	return result, nil
}

// GetRuntimeMutatorTargets returns runtime mutator targets for the given system.
// Not yet backed by persisted metadata; returns empty so chaos-experiment's
// routing layer falls back to bundled defaults. When/if we persist mutator
// targets, mirror the GetNetworkPairs pattern using a new meta type.
func (s *DBMetadataStore) GetRuntimeMutatorTargets(system string) ([]chaos.RuntimeMutatorTargetData, error) {
	return nil, nil
}

// InvalidateCache clears the entire cache. Call after metadata updates.
func (s *DBMetadataStore) InvalidateCache() {
	s.cache.Range(func(key, _ interface{}) bool {
		s.cache.Delete(key)
		return true
	})
}
