# ContainerVersionDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **string** |  | [optional] [default to undefined]
**env_vars** | **string** |  | [optional] [default to undefined]
**github_link** | **string** |  | [optional] [default to undefined]
**helm_config** | [**HelmConfigDetailResp**](HelmConfigDetailResp.md) |  | [optional] [default to undefined]
**id** | **number** |  | [optional] [default to undefined]
**image_ref** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**updated_at** | **string** |  | [optional] [default to undefined]
**usage** | **number** |  | [optional] [default to undefined]

## Example

```typescript
import { ContainerVersionDetailResp } from 'rcabench-client';

const instance: ContainerVersionDetailResp = {
    command,
    env_vars,
    github_link,
    helm_config,
    id,
    image_ref,
    name,
    updated_at,
    usage,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
