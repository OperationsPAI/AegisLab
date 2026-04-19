package pedestal

import (
	"errors"
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/middleware"
	"aegis/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	repo   *Repository
	runner Runner
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo, runner: RealRunner{}}
}

// GetPedestalHelmConfig returns the helm_configs row for a given container_version_id.
//
//	@Summary		Get pedestal helm config
//	@Description	Retrieve the helm chart configuration bound to a pedestal container version.
//	@Tags			Pedestal
//	@ID				get_pedestal_helm_config
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_version_id	path		int										true	"Container version ID"
//	@Success		200						{object}	dto.GenericResponse[PedestalHelmConfigResp]
//	@Failure		400						{object}	dto.GenericResponse[any]
//	@Failure		401						{object}	dto.GenericResponse[any]
//	@Failure		404						{object}	dto.GenericResponse[any]
//	@Router			/api/v2/pedestal/helm/{container_version_id} [get]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) GetPedestalHelmConfig(c *gin.Context) {
	if _, ok := middleware.GetCurrentUserID(c); !ok {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	versionID, ok := parseVersionID(c)
	if !ok {
		return
	}

	cfg, err := h.repo.GetHelmConfigByContainerVersionID(versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dto.ErrorResponse(c, http.StatusNotFound, "Helm config not found for container_version_id")
			return
		}
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load helm config: "+err.Error())
		return
	}

	dto.SuccessResponse(c, toHelmConfigResp(cfg))
}

// UpsertPedestalHelmConfig creates or updates the helm_configs row for the given container version.
//
//	@Summary		Upsert pedestal helm config
//	@Description	Create or update the helm_configs row for a pedestal container version. Admin-only.
//	@Tags			Pedestal
//	@ID				upsert_pedestal_helm_config
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_version_id	path	int								true	"Container version ID"
//	@Param			request					body	UpsertPedestalHelmConfigReq	true	"Helm config fields"
//	@Success		200	{object}	dto.GenericResponse[PedestalHelmConfigResp]
//	@Router			/api/v2/pedestal/helm/{container_version_id} [put]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) UpsertPedestalHelmConfig(c *gin.Context) {
	if _, ok := middleware.GetCurrentUserID(c); !ok {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	versionID, ok := parseVersionID(c)
	if !ok {
		return
	}

	var req UpsertPedestalHelmConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	fresh, err := h.repo.UpsertHelmConfig(versionID, &model.HelmConfig{
		ChartName: req.ChartName,
		Version:   req.Version,
		RepoURL:   req.RepoURL,
		RepoName:  req.RepoName,
		ValueFile: req.ValueFile,
		LocalPath: req.LocalPath,
	})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to upsert helm config: "+err.Error())
		return
	}

	dto.SuccessResponse(c, toHelmConfigResp(fresh))
}

// VerifyPedestalHelmConfig dry-runs helm repo add + helm pull + value-file parse.
//
//	@Summary		Verify pedestal helm config
//	@Description	Dry-run helm repo add + pull and parse the values file without starting a task.
//	@Tags			Pedestal
//	@ID				verify_pedestal_helm_config
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_version_id	path	int	true	"Container version ID"
//	@Success		200	{object}	dto.GenericResponse[PedestalHelmVerifyResp]
//	@Router			/api/v2/pedestal/helm/{container_version_id}/verify [post]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) VerifyPedestalHelmConfig(c *gin.Context) {
	if _, ok := middleware.GetCurrentUserID(c); !ok {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	versionID, ok := parseVersionID(c)
	if !ok {
		return
	}

	cfg, err := h.repo.GetHelmConfigByContainerVersionID(versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dto.ErrorResponse(c, http.StatusNotFound, "Helm config not found for container_version_id")
			return
		}
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load helm config: "+err.Error())
		return
	}

	result := Run(h.runner, Config{
		ChartName: cfg.ChartName,
		Version:   cfg.Version,
		RepoURL:   cfg.RepoURL,
		RepoName:  cfg.RepoName,
		ValueFile: cfg.ValueFile,
	}, VerifyValueFile)

	resp := PedestalHelmVerifyResp{OK: result.OK, Checks: make([]PedestalHelmVerifyCheck, len(result.Checks))}
	for i, chk := range result.Checks {
		resp.Checks[i] = PedestalHelmVerifyCheck{Name: chk.Name, OK: chk.OK, Detail: chk.Detail}
	}
	dto.SuccessResponse(c, resp)
}

func parseVersionID(c *gin.Context) (int, bool) {
	raw := c.Param("container_version_id")
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container_version_id: "+raw)
		return 0, false
	}
	return id, true
}

func toHelmConfigResp(cfg *model.HelmConfig) PedestalHelmConfigResp {
	return PedestalHelmConfigResp{
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
