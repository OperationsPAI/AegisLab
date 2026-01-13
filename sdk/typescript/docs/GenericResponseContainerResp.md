# GenericResponseContainerResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **number** | Status code | [optional] [default to undefined]
**data** | [**ContainerResp**](ContainerResp.md) | Generic type data | [optional] [default to undefined]
**message** | **string** | Response message | [optional] [default to undefined]
**timestamp** | **number** | Response generation time | [optional] [default to undefined]

## Example

```typescript
import { GenericResponseContainerResp } from 'rcabench-client';

const instance: GenericResponseContainerResp = {
    code,
    data,
    message,
    timestamp,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
