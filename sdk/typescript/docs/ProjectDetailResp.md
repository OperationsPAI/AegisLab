# ProjectDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**containers** | [**Array&lt;ContainerResp&gt;**](ContainerResp.md) |  | [optional] [default to undefined]
**created_at** | **string** |  | [optional] [default to undefined]
**datapacks** | [**Array&lt;InjectionResp&gt;**](InjectionResp.md) |  | [optional] [default to undefined]
**datasets** | [**Array&lt;DatasetResp&gt;**](DatasetResp.md) |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**id** | **number** |  | [optional] [default to undefined]
**is_public** | **boolean** |  | [optional] [default to undefined]
**labels** | [**Array&lt;LabelItem&gt;**](LabelItem.md) |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**status** | **string** |  | [optional] [default to undefined]
**updated_at** | **string** |  | [optional] [default to undefined]
**user_count** | **number** |  | [optional] [default to undefined]

## Example

```typescript
import { ProjectDetailResp } from 'rcabench-client';

const instance: ProjectDetailResp = {
    containers,
    created_at,
    datapacks,
    datasets,
    description,
    id,
    is_public,
    labels,
    name,
    status,
    updated_at,
    user_count,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
