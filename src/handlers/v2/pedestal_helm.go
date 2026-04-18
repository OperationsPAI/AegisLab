package v2

import (
	"errors"
	"net/http"
	"strconv"

	"aegis/database"
	"aegis/dto"
	"aegis/handlers/v2/pedestalhelm"
	"aegis/middleware"
	"aegis/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// defaultHelmRunner is overridable from tests / future server wiring.
var defaultHelmRunner pedestalhelm.Runner = pedestalhelm.RealRunner{}

// GetPedestalHelmConfig returns the helm_configs row for a given container_version_id.
//
//	@Summary		Get pedestal helm config
//	@Description	Retrieve the helm chart configuration bound to a pedestal container version.
//	@Tags			Pedestal
//	@ID				get_pedestal_helm_config
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_version_id	path		int										true	"Container version ID"
//	@Success		200						{object}	dto.GenericResponse[dto.PedestalHelmConfigResp]
//	@Failure		400						{object}	dto.GenericResponse[any]
//	@Failure		401						{object}	dto.GenericResponse[any]
//	@Failure		404						{object}	dto.GenericResponse[any]
//	@Router			/api/v2/pedestal/helm/{container_version_id} [get]
//	@x-api-type		{"sdk":"true"}
func GetPedestalHelmConfig(c *gin.Context) {
	if _, ok := middleware.GetCurrentUserID(c); !ok {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	versionID, ok := parsePedestalVersionID(c)
	if !ok {
		return
	}

	cfg, err := repository.GetHelmConfigByContainerVersionID(database.DB, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dto.ErrorResponse(c, http.StatusNotFound, "helm config not found for container_version_id")
			return
		}
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to load helm config: "+err.Error())
		return
	}

	dto.SuccessResponse(c, toHelmConfigResp(cfg))
}

// UpsertPedestalHelmConfig creates or updates the helm_configs row for the
// given container version. Requires container-version upload permission
// (same tier as POST .../helm-chart).
//
//	@Summary		Upsert pedestal helm config
//	@Description	Create or update the helm_configs row for a pedestal container version. Admin-only.
//	@Tags			Pedestal
//	@ID				upsert_pedestal_helm_config
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_version_id	path	int								true	"Container version ID"
//	@Param			request					body	dto.UpsertPedestalHelmConfigReq	true	"Helm config fields"
//	@Success		200	{object}	dto.GenericResponse[dto.PedestalHelmConfigResp]
//	@Router			/api/v2/pedestal/helm/{container_version_id} [put]
//	@x-api-type		{"sdk":"true"}
func UpsertPedestalHelmConfig(c *gin.Context) {
	if _, ok := middleware.GetCurrentUserID(c); !ok {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	versionID, ok := parsePedestalVersionID(c)
	if !ok {
		return
	}

	var req dto.UpsertPedestalHelmConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	existing, err := repository.GetHelmConfigByContainerVersionID(database.DB, versionID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to query existing helm config: "+err.Error())
		return
	}

	if existing != nil && existing.ID != 0 {
		existing.ChartName = req.ChartName
		existing.Version = req.Version
		existing.RepoURL = req.RepoURL
		existing.RepoName = req.RepoName
		existing.ValueFile = req.ValueFile
		existing.LocalPath = req.LocalPath
		if err := repository.UpdateHelmConfig(database.DB, existing); err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to update helm config: "+err.Error())
			return
		}
		dto.SuccessResponse(c, toHelmConfigResp(existing))
		return
	}

	created := &database.HelmConfig{
		ContainerVersionID: versionID,
		ChartName:          req.ChartName,
		Version:            req.Version,
		RepoURL:            req.RepoURL,
		RepoName:           req.RepoName,
		ValueFile:          req.ValueFile,
		LocalPath:          req.LocalPath,
	}
	if err := repository.BatchCreateHelmConfigs(database.DB, []*database.HelmConfig{created}); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to create helm config: "+err.Error())
		return
	}
	dto.SuccessResponse(c, toHelmConfigResp(created))
}

// VerifyPedestalHelmConfig dry-runs helm repo add + helm pull + value-file
// parse. It never triggers a real restart_pedestal task.
//
//	@Summary		Verify pedestal helm config
//	@Description	Dry-run helm repo add + pull and parse the values file without starting a task.
//	@Tags			Pedestal
//	@ID				verify_pedestal_helm_config
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_version_id	path	int	true	"Container version ID"
//	@Success		200	{object}	dto.GenericResponse[dto.PedestalHelmVerifyResp]
//	@Router			/api/v2/pedestal/helm/{container_version_id}/verify [post]
//	@x-api-type		{"sdk":"true"}
func VerifyPedestalHelmConfig(c *gin.Context) {
	if _, ok := middleware.GetCurrentUserID(c); !ok {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	versionID, ok := parsePedestalVersionID(c)
	if !ok {
		return
	}

	cfg, err := repository.GetHelmConfigByContainerVersionID(database.DB, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dto.ErrorResponse(c, http.StatusNotFound, "helm config not found for container_version_id")
			return
		}
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to load helm config: "+err.Error())
		return
	}

	result := pedestalhelm.Run(defaultHelmRunner, pedestalhelm.Config{
		ChartName: cfg.ChartName,
		Version:   cfg.Version,
		RepoURL:   cfg.RepoURL,
		RepoName:  cfg.RepoName,
		ValueFile: cfg.ValueFile,
	}, pedestalhelm.VerifyValueFile)

	resp := dto.PedestalHelmVerifyResp{OK: result.OK, Checks: make([]dto.PedestalHelmVerifyCheck, len(result.Checks))}
	for i, chk := range result.Checks {
		resp.Checks[i] = dto.PedestalHelmVerifyCheck{Name: chk.Name, OK: chk.OK, Detail: chk.Detail}
	}
	dto.SuccessResponse(c, resp)
}

func parsePedestalVersionID(c *gin.Context) (int, bool) {
	raw := c.Param("container_version_id")
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "invalid container_version_id: "+raw)
		return 0, false
	}
	return id, true
}

func toHelmConfigResp(cfg *database.HelmConfig) dto.PedestalHelmConfigResp {
	return dto.PedestalHelmConfigResp{
		ID:                 cfg.ID,
		ContainerVersionID: cfg.ContainerVersionID,
		ChartName:          cfg.ChartName,
		Version:            cfg.Version,
		RepoURL:            cfg.RepoURL,
		RepoName:           cfg.RepoName,
		ValueFile:          cfg.ValueFile,
		LocalPath:          cfg.LocalPath,
		Checksum:           cfg.Checksum,
	}
}
