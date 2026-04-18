package taskmodule

import (
	"context"

	"aegis/dto"
	"aegis/model"

	"github.com/gorilla/websocket"
)

// HandlerService captures task operations consumed by HTTP handlers and gateway adapters.
type HandlerService interface {
	BatchDelete(context.Context, []string) error
	GetDetail(context.Context, string) (*TaskDetailResp, error)
	List(context.Context, *ListTaskReq) (*dto.ListResp[TaskResp], error)
	GetForLogStream(context.Context, string) (*model.Task, error)
	StreamLogs(context.Context, *websocket.Conn, *model.Task)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
