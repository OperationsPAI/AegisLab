# SearchDatasetReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | [**DateRange**](DateRange.md) |  | [optional] [default to undefined]
**include_versions** | **boolean** |  | [optional] [default to undefined]
**name_pattern** | **string** |  | [optional] [default to undefined]
**page** | **number** |  | [optional] [default to undefined]
**size** | [**PageSize**](PageSize.md) |  | [optional] [default to undefined]
**sort** | [**Array&lt;SortOption&gt;**](SortOption.md) |  | [optional] [default to undefined]
**status** | [**Array&lt;StatusType&gt;**](StatusType.md) |  | [optional] [default to undefined]
**updated_at** | [**DateRange**](DateRange.md) |  | [optional] [default to undefined]

## Example

```typescript
import { SearchDatasetReq } from 'rcabench-client';

const instance: SearchDatasetReq = {
    created_at,
    include_versions,
    name_pattern,
    page,
    size,
    sort,
    status,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
