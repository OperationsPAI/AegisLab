package consts

type DatapackStateString string

const (
	DatapackInitialName         DatapackStateString = "initial"
	DatapackInjectFailedName    DatapackStateString = "inject_failed"
	DatapackInjectSuccessName   DatapackStateString = "inject_success"
	DatapackBuildFailedName     DatapackStateString = "build_failed"
	DatapackBuildSuccessName    DatapackStateString = "build_success"
	DatapackDetectorFailedName  DatapackStateString = "detector_failed"
	DatapackDetectorSuccessName DatapackStateString = "detector_success"
)

type ExecutionStateString string

const (
	ExecutionInitialName ExecutionStateString = "Initial"
	ExecutionFailedName  ExecutionStateString = "Failed"
	ExecutionSuccessName ExecutionStateString = "Success"
)
