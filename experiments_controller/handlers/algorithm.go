package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

// GetAlgorithmResp
// 去获取每个算法目录里的 toml 描述文件
type GetAlgorithmResp struct {
	Name string `json:"name"`
}

// GetAlgorithms
//
//	@Summary		获取算法列表
//	@Description	获取算法列表
//	@Tags			algorithm
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		InjectReq	true	"请求体"
//	@Success		200		{object}	GenericResponse[GetAlgorithmResp]
//	@Failure		400		{object}	GenericResponse[GetAlgorithmResp]
//	@Failure		500		{object}	GenericResponse[GetAlgorithmResp]
//	@Router			/api/v1/algo/injectstatus [post]
func GetAlgorithms(c *gin.Context) {
	pwd, err := os.Getwd()
	if err != nil {
		c.JSON(500, gin.H{"error": "Get work directory failed"})
		return
	}

	parentDir := filepath.Dir(pwd)
	algoPath := filepath.Join(parentDir, "algorithms")

	algoFiles, err := getSubFiles(algoPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list files in %s: %v", algoPath, err)})
		return
	}

	var algoResps []GetAlgorithmResp
	for _, algoFile := range algoFiles {
		tomlPath := filepath.Join(algoPath, algoFile, "info.toml")
		var algoResp GetAlgorithmResp
		if _, err := toml.DecodeFile(tomlPath, &algoResp); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get info.toml in %s: %v", algoPath, err)})
			return
		}

		algoResps = append(algoResps, algoResp)
	}

	c.JSON(http.StatusOK, gin.H{"algorithms": algoResps})
}
