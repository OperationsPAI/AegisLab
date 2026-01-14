# UserDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avatar** | **string** |  | [optional] [default to undefined]
**container_roles** | [**Array&lt;UserContainerInfo&gt;**](UserContainerInfo.md) |  | [optional] [default to undefined]
**created_at** | **string** |  | [optional] [default to undefined]
**dataset_roles** | [**Array&lt;UserDatasetInfo&gt;**](UserDatasetInfo.md) |  | [optional] [default to undefined]
**email** | **string** |  | [optional] [default to undefined]
**full_name** | **string** |  | [optional] [default to undefined]
**global_roles** | [**Array&lt;RoleResp&gt;**](RoleResp.md) |  | [optional] [default to undefined]
**id** | **number** |  | [optional] [default to undefined]
**is_active** | **boolean** |  | [optional] [default to undefined]
**last_login_at** | **string** |  | [optional] [default to undefined]
**permissions** | [**Array&lt;PermissionResp&gt;**](PermissionResp.md) |  | [optional] [default to undefined]
**phone** | **string** |  | [optional] [default to undefined]
**project_roles** | [**Array&lt;UserProjectInfo&gt;**](UserProjectInfo.md) |  | [optional] [default to undefined]
**status** | **string** |  | [optional] [default to undefined]
**updated_at** | **string** |  | [optional] [default to undefined]
**username** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { UserDetailResp } from 'rcabench-client';

const instance: UserDetailResp = {
    avatar,
    container_roles,
    created_at,
    dataset_roles,
    email,
    full_name,
    global_roles,
    id,
    is_active,
    last_login_at,
    permissions,
    phone,
    project_roles,
    status,
    updated_at,
    username,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
