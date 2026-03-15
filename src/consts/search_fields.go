package consts

// InjectionField represents a valid field name for injection sort/group_by operations.
// Only fields defined here are accepted; any other value will be rejected to prevent SQL injection.
type InjectionField string

const (
	InjectionFieldName      InjectionField = "name"
	InjectionFieldFaultType InjectionField = "fault_type"
	InjectionFieldCategory  InjectionField = "category"
	InjectionFieldState     InjectionField = "state"
	InjectionFieldStartTime InjectionField = "start_time"
	InjectionFieldEndTime   InjectionField = "end_time"
	InjectionFieldCreatedAt InjectionField = "created_at"
	InjectionFieldUpdatedAt InjectionField = "updated_at"
)

func (f InjectionField) String() string { return string(f) }

// InjectionAllowedFields is the authoritative whitelist for injection sort and group_by fields.
// Key: user-facing field name, Value: actual database column name.
var InjectionAllowedFields = map[InjectionField]string{
	InjectionFieldName:      "name",
	InjectionFieldFaultType: "fault_type",
	InjectionFieldCategory:  "category",
	InjectionFieldState:     "state",
	InjectionFieldStartTime: "start_time",
	InjectionFieldEndTime:   "end_time",
	InjectionFieldCreatedAt: "created_at",
	InjectionFieldUpdatedAt: "updated_at",
}

// DatasetField represents a valid field name for dataset sort/group_by operations.
// Only fields defined here are accepted; any other value will be rejected to prevent SQL injection.
type DatasetField string

const (
	DatasetFieldName      DatasetField = "name"
	DatasetFieldIsPublic  DatasetField = "is_public"
	DatasetFieldCreatedAt DatasetField = "created_at"
	DatasetFieldUpdatedAt DatasetField = "updated_at"
)

func (f DatasetField) String() string { return string(f) }

// DatasetAllowedFields is the authoritative whitelist for dataset sort and group_by fields.
// Key: user-facing field name, Value: actual database column name.
var DatasetAllowedFields = map[DatasetField]string{
	DatasetFieldName:      "name",
	DatasetFieldIsPublic:  "is_public",
	DatasetFieldCreatedAt: "created_at",
	DatasetFieldUpdatedAt: "updated_at",
}
