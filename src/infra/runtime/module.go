package runtimeinfra

import (
	"time"

	"aegis/consts"
	"aegis/utils"

	"go.uber.org/fx"
)

var Module = fx.Module("runtime",
	fx.Invoke(InitializeRuntime),
)

func InitializeRuntime() {
	if consts.InitialTime == nil {
		consts.InitialTime = utils.TimePtr(time.Now())
	}
	if consts.AppID == "" {
		consts.AppID = utils.GenerateULID(consts.InitialTime)
	}
}
