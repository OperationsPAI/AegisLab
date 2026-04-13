package producer

import (
	"fmt"
	"regexp"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
)

// ListChaosSystemsService lists chaos systems with pagination
func ListChaosSystemsService(req *dto.ListChaosSystemReq) (*dto.ListResp[dto.ChaosSystemResp], error) {
	limit, offset := req.ToGormParams()

	systems, total, err := repository.ListSystems(database.DB, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}

	items := make([]dto.ChaosSystemResp, 0, len(systems))
	for _, s := range systems {
		items = append(items, *dto.NewChaosSystemResp(&s))
	}

	return &dto.ListResp[dto.ChaosSystemResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

// GetChaosSystemService retrieves a single chaos system by ID
func GetChaosSystemService(id int) (*dto.ChaosSystemResp, error) {
	system, err := repository.GetSystemByID(database.DB, id)
	if err != nil {
		return nil, err
	}
	return dto.NewChaosSystemResp(system), nil
}

// CreateChaosSystemService creates a new chaos system and registers it with chaos-experiment
func CreateChaosSystemService(req *dto.CreateChaosSystemReq) (*dto.ChaosSystemResp, error) {
	// Validate regex patterns
	if _, err := regexp.Compile(req.NsPattern); err != nil {
		return nil, fmt.Errorf("invalid ns_pattern regex: %w: %w", err, consts.ErrBadRequest)
	}
	if _, err := regexp.Compile(req.ExtractPattern); err != nil {
		return nil, fmt.Errorf("invalid extract_pattern regex: %w: %w", err, consts.ErrBadRequest)
	}

	system := &database.System{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		NsPattern:      req.NsPattern,
		ExtractPattern: req.ExtractPattern,
		Count:          req.Count,
		Description:    req.Description,
		IsBuiltin:      false,
		Status:         consts.CommonEnabled,
	}

	if err := repository.CreateSystem(database.DB, system); err != nil {
		return nil, fmt.Errorf("failed to create system: %w", err)
	}

	// Register with chaos-experiment
	if err := chaos.RegisterSystem(chaos.SystemConfig{
		Name:        system.Name,
		NsPattern:   system.NsPattern,
		DisplayName: system.DisplayName,
	}); err != nil {
		logrus.WithError(err).Warnf("Failed to register system %s with chaos-experiment", system.Name)
	}

	return dto.NewChaosSystemResp(system), nil
}

// UpdateChaosSystemService updates a chaos system and re-registers it
func UpdateChaosSystemService(id int, req *dto.UpdateChaosSystemReq) (*dto.ChaosSystemResp, error) {
	system, err := repository.GetSystemByID(database.DB, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.NsPattern != nil {
		if _, err := regexp.Compile(*req.NsPattern); err != nil {
			return nil, fmt.Errorf("invalid ns_pattern regex: %w: %w", err, consts.ErrBadRequest)
		}
		updates["ns_pattern"] = *req.NsPattern
	}
	if req.ExtractPattern != nil {
		if _, err := regexp.Compile(*req.ExtractPattern); err != nil {
			return nil, fmt.Errorf("invalid extract_pattern regex: %w: %w", err, consts.ErrBadRequest)
		}
		updates["extract_pattern"] = *req.ExtractPattern
	}
	if req.Count != nil {
		updates["count"] = *req.Count
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		return dto.NewChaosSystemResp(system), nil
	}

	if err := repository.UpdateSystem(database.DB, id, updates); err != nil {
		return nil, err
	}

	// Reload the system to get updated fields
	system, err = repository.GetSystemByID(database.DB, id)
	if err != nil {
		return nil, err
	}

	// Re-register with chaos-experiment
	if err := chaos.RegisterSystem(chaos.SystemConfig{
		Name:        system.Name,
		NsPattern:   system.NsPattern,
		DisplayName: system.DisplayName,
	}); err != nil {
		logrus.WithError(err).Warnf("Failed to re-register system %s with chaos-experiment", system.Name)
	}

	return dto.NewChaosSystemResp(system), nil
}

// DeleteChaosSystemService soft-deletes a chaos system
func DeleteChaosSystemService(id int) error {
	system, err := repository.GetSystemByID(database.DB, id)
	if err != nil {
		return err
	}

	if system.IsBuiltin {
		return fmt.Errorf("cannot delete builtin system %s: %w", system.Name, consts.ErrBadRequest)
	}

	if err := repository.DeleteSystem(database.DB, id); err != nil {
		return err
	}

	// Unregister from chaos-experiment
	if err := chaos.UnregisterSystem(system.Name); err != nil {
		logrus.WithError(err).Warnf("Failed to unregister system %s from chaos-experiment", system.Name)
	}

	return nil
}

// UpsertChaosSystemMetadataService bulk upserts metadata for a system
func UpsertChaosSystemMetadataService(id int, req *dto.BulkUpsertSystemMetadataReq) error {
	system, err := repository.GetSystemByID(database.DB, id)
	if err != nil {
		return err
	}

	for _, item := range req.Items {
		meta := &database.SystemMetadata{
			SystemName:   system.Name,
			MetadataType: item.MetadataType,
			ServiceName:  item.ServiceName,
			Data:         string(item.Data),
		}
		if err := repository.UpsertSystemMetadata(database.DB, meta); err != nil {
			return fmt.Errorf("failed to upsert metadata (type=%s, service=%s): %w", item.MetadataType, item.ServiceName, err)
		}
	}

	return nil
}

// ListChaosSystemMetadataService lists metadata for a system, optionally filtered by type
func ListChaosSystemMetadataService(id int, metadataType string) ([]dto.SystemMetadataResp, error) {
	system, err := repository.GetSystemByID(database.DB, id)
	if err != nil {
		return nil, err
	}

	metas, err := repository.ListSystemMetadata(database.DB, system.Name, metadataType)
	if err != nil {
		return nil, fmt.Errorf("failed to list system metadata: %w", err)
	}

	items := make([]dto.SystemMetadataResp, 0, len(metas))
	for _, m := range metas {
		items = append(items, *dto.NewSystemMetadataResp(&m))
	}

	return items, nil
}
