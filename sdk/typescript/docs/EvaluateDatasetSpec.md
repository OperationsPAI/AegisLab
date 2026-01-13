# EvaluateDatasetSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | [**ContainerRef**](ContainerRef.md) |  | [default to undefined]
**dataset** | [**DatasetRef**](DatasetRef.md) |  | [default to undefined]
**filter_labels** | [**Array&lt;LabelItem&gt;**](LabelItem.md) |  | [optional] [default to undefined]

## Example

```typescript
import { EvaluateDatasetSpec } from 'rcabench-client';

const instance: EvaluateDatasetSpec = {
    algorithm,
    dataset,
    filter_labels,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
