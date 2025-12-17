package common

import (
	"aegis/consts"

	"github.com/sirupsen/logrus"
)

var TraceTaskEventMap = map[consts.TaskType]map[consts.TaskState]consts.EventType{
	consts.TaskTypeRestartPedestal: {
		consts.TaskRunning:     consts.EventRestartPedestalStarted,
		consts.TaskCompleted:   consts.EventRestartPedestalCompleted,
		consts.TaskError:       consts.EventRestartPedestalFailed,
		consts.TaskRescheduled: consts.EventNoNamespaceAvailable,
	},
	consts.TaskTypeFaultInjection: {
		consts.TaskRunning:   consts.EventFaultInjectionStarted,
		consts.TaskCompleted: consts.EventFaultInjectionCompleted,
		consts.TaskError:     consts.EventFaultInjectionFailed,
	},
	consts.TaskTypeBuildDatapack: {
		consts.TaskRunning:   consts.EventDatapackBuildStarted,
		consts.TaskCompleted: consts.EventDatapackBuildSucceed,
		consts.TaskError:     consts.EventDatapackBuildFailed,
	},
	consts.TaskTypeRunAlgorithm: {
		consts.TaskRunning:     consts.EventAlgoRunStarted,
		consts.TaskCompleted:   consts.EventAlgoRunSucceed,
		consts.TaskError:       consts.EventAlgoRunFailed,
		consts.TaskRescheduled: consts.EventNoTokenAvailable,
	},
}

// GetEventTypeByTask maps a task type and state to the corresponding event type
func GetEventTypeByTask(taskType consts.TaskType, taskState consts.TaskState) consts.EventType {
	stateMap, exists := TraceTaskEventMap[taskType]
	if !exists {
		logrus.Warnf("no event type mapping for task type: %s", consts.GetTaskTypeName(taskType))
		return "unknown"
	}

	eventType, exists := stateMap[taskState]
	if !exists {
		logrus.Warnf("no event type mapping for task state: %s", consts.GetTaskStateName(taskState))
		return "unknown"
	}

	return eventType
}
