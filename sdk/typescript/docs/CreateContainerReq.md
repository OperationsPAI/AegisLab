# CreateContainerReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**is_public** | **boolean** |  | [optional] [default to undefined]
**name** | **string** |  | [default to undefined]
**readme** | **string** |  | [optional] [default to undefined]
**type** | [**ContainerType**](ContainerType.md) |  | [optional] [default to undefined]
**version** | [**CreateContainerVersionReq**](CreateContainerVersionReq.md) |  | [optional] [default to undefined]

## Example

```typescript
import { CreateContainerReq } from 'rcabench-client';

const instance: CreateContainerReq = {
    is_public,
    name,
    readme,
    type,
    version,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
