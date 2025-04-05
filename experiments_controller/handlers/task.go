package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetTaskDetail(c *gin.Context) {
	var req dto.TaskReq
	if err := c.BindUri(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	logEntry := logrus.WithField("task_id", req.TaskID)

	taskItem, err := repository.FindTaskByID(req.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
		} else {
			message := "failed to retrieve task of injection"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
		}

		return
	}

	logKey := fmt.Sprintf("task:%s:logs", taskItem.ID)
	logs, err := client.GetRedisClient().LRange(c.Request.Context(), logKey, 0, -1).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			message := "Failed to retrieve logs"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		logs = []string{}
	}

	dto.SuccessResponse(c, dto.TaskDetailResp{Task: *taskItem, Logs: logs})
}
