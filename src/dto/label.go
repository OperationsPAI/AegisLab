package dto

type LabelItem struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSystem bool   `json:"is_system,omitempty"`
}

// ConvertLabelItemssToConditions converts a slice of LabelItem to a slice of map conditions
func ConvertLabelItemsToConditions(labelItems []LabelItem) []map[string]string {
	if len(labelItems) == 0 {
		return []map[string]string{}
	}

	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, label := range labelItems {
		labelConditions = append(labelConditions, map[string]string{
			"key":   label.Key,
			"value": label.Value,
		})
	}

	return labelConditions
}
