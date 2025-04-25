package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	conf "github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func CancelInjection(c *gin.Context) {
}

func GetInjectionConf(c *gin.Context) {
	var req dto.InjectionConfReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	root, err := chaos.StructToNode[handler.InjectionConf]()
	if err != nil {
		logrus.Errorf("struct InjectionConf to node failed: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read injection conf")
		return
	}

	if req.Mode == "engine" {
		dto.SuccessResponse(c, chaos.NodeToMap(root, false))
		return
	}

	type NodeItem struct {
		Description string `json:"description"`
		Range       []int  `json:"range"`
	}

	type result struct {
		key   string
		value map[string]NodeItem
	}

	chaosMap := make(map[string]map[string]NodeItem, len(root.Children))
	resultChan := make(chan result, len(root.Children))
	var wg sync.WaitGroup

	// 并行处理每个节点
	for _, node := range root.Children {
		wg.Add(1)
		go func(n *chaos.Node) {
			defer wg.Done()
			m := make(map[string]NodeItem, len(n.Children))
			for _, child := range n.Children {
				m[child.Name] = NodeItem{
					Description: child.Description,
					Range:       child.Range,
				}
			}
			resultChan <- result{key: n.Name, value: m}
		}(node)
	}

	// 等待所有处理完成并关闭channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集处理结果
	for res := range resultChan {
		chaosMap[res.key] = res.value
	}

	dto.SuccessResponse(c, chaosMap)
}

// GetInjectionList
//
//	@Summary		分页查询注入记录列表
//	@Description	获取注入记录列表（支持分页参数）
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Success		200	{object}		dto.GenericResponse[dto.PaginationResp[dto.InjectionItem]]
//	@Failure		400	{object}		dto.GenericResponse[any]
//	@Failure		500	{object}		dto.GenericResponse[any]
//	@Router			/api/v1/injections/getlist [post]
func GetInjectionList(c *gin.Context) {
	var req dto.InjectionListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, map[string]string{}))
		return
	}

	total, records, err := repository.ListInjectionWithPagination(req.PageNum, req.PageSize)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]dto.InjectionItem, 0, len(records))
	for _, record := range records {
		var item dto.InjectionItem
		if err := item.Convert(record); err != nil {
			logrus.WithField("injection", record.ID).Error(err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "invalid injection configuration")
			return
		}

		items = append(items, item)
	}

	dto.SuccessResponse(c, &dto.PaginationResp[dto.InjectionItem]{
		Total: total,
		Data:  items,
	})
}

// SubmitFaultInjection
//
//	@Summary		注入故障
//	@Description	注入故障
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		[]dto.InjectionSubmitReq	true	"请求体"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections [post]
func SubmitFaultInjection(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitFaultInjection called, groupID: %s", groupID)

	var req dto.InjectionSubmitReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	configs, err := req.ParseInjectionSpecs()
	if err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if !conf.GetBool("injection.enable_duplicate") {
		newConfigs, err := getNewConfigs(configs, req.Interval)
		if err != nil {
			message := "failed to get the existing configs"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		configs = newConfigs
	}

	ctx, span := otel.Tracer("rcabench/group").Start(context.Background(), "produce group", trace.WithAttributes(
		attribute.String("group_id", groupID),
	))
	defer span.End()

	traces := make([]dto.Trace, 0, len(configs))
	for _, config := range configs {
		payload := map[string]any{
			consts.RestartIntarval:      req.Interval,
			consts.RestartFaultDuration: config.FaultDuration,
			consts.RestartInjectPayload: map[string]any{
				consts.InjectBenchmark:   req.Benchmark,
				consts.InjectFaultType:   config.FaultType,
				consts.InjectPreDuration: req.PreDuration,
				consts.InjectDisplayData: config.DisplayData,
				consts.InjectConf:        config.Conf,
				consts.InjectNode:        config.Node,
			},
		}

		task := &executor.UnifiedTask{
			Type:        consts.TaskTypeRestartService,
			Payload:     payload,
			Immediate:   false,
			ExecuteTime: config.ExecuteTime.Unix(),
			GroupID:     groupID,
		}
		task.GroupCarrier = make(propagation.MapCarrier)
		otel.GetTextMapPropagator().Inject(ctx, task.GroupCarrier)

		taskID, traceID, err := executor.SubmitTask(context.Background(), task)
		if err != nil {
			message := "failed to submit injection task"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: config.Index})
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(traces), groupID))

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", dto.SubmitResp{GroupID: groupID, Traces: traces})
}

func getNewConfigs(configs []*dto.InjectionConfig, interval int) ([]*dto.InjectionConfig, error) {
	intervalDuration := time.Duration(interval) * consts.DefaultTimeUnit

	displayDatas := make([]string, 0, len(configs))
	for _, config := range configs {
		displayDatas = append(displayDatas, config.DisplayData)
	}

	missingIndices, err := findMissingIndices(displayDatas, 10)
	if err != nil {
		return nil, err
	}

	logrus.Infof("deduplicated %d configurations (remaining: %d)", len(displayDatas)-len(missingIndices), len(missingIndices))

	newConfigs := make([]*dto.InjectionConfig, 0, len(missingIndices))
	for i, idx := range missingIndices {
		config := configs[idx]
		config.ExecuteTime = config.ExecuteTime.Add(-intervalDuration * time.Duration(config.Index-i))
		newConfigs = append(newConfigs, config)
	}

	return newConfigs, nil
}

func findMissingIndices(confs []string, batch_size int) ([]int, error) {
	var missingIndices []int
	existingMap := make(map[string]struct{})

	for i := 0; i < len(confs); i += batch_size {
		end := min(i+batch_size, len(confs))

		batch := confs[i:end]
		existingBatch, err := repository.FindExistingDisplayConfigs(batch)
		if err != nil {
			return nil, err
		}

		for _, s := range existingBatch {
			existingMap[s] = struct{}{}
		}
	}

	for idx, s := range confs {
		if _, exists := existingMap[s]; !exists {
			missingIndices = append(missingIndices, idx)
		}
	}

	return missingIndices, nil
}
