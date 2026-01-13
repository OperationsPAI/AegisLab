# HelmConfigItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**chart_name** | **string** |  | [optional] [default to undefined]
**full_chart** | **string** |  | [optional] [default to undefined]
**repo_name** | **string** |  | [optional] [default to undefined]
**repo_url** | **string** |  | [optional] [default to undefined]
**value_file** | **string** |  | [optional] [default to undefined]
**values** | [**Array&lt;ParameterItem&gt;**](ParameterItem.md) |  | [optional] [default to undefined]

## Example

```typescript
import { HelmConfigItem } from 'rcabench-client';

const instance: HelmConfigItem = {
    chart_name,
    full_chart,
    repo_name,
    repo_url,
    value_file,
    values,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
