# EvaluateDatasetItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **string** | Algorithm name | [optional] [default to undefined]
**algorithm_version** | **string** | Algorithm version | [optional] [default to undefined]
**dataset** | **string** | Dataset name | [optional] [default to undefined]
**dataset_version** | **string** | Dataset version | [optional] [default to undefined]
**evalaute_refs** | [**Array&lt;EvaluateDatapackRef&gt;**](EvaluateDatapackRef.md) | Evaluation refs for each dataset | [optional] [default to undefined]
**not_executed_datapacks** | **Array&lt;string&gt;** | Datapacks that were not executed | [optional] [default to undefined]
**total_count** | **number** | Total number of datapacks in dataset | [optional] [default to undefined]

## Example

```typescript
import { EvaluateDatasetItem } from 'rcabench-client';

const instance: EvaluateDatasetItem = {
    algorithm,
    algorithm_version,
    dataset,
    dataset_version,
    evalaute_refs,
    not_executed_datapacks,
    total_count,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
