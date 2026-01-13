# CreateParameterConfigReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | [**ParameterCategory**](ParameterCategory.md) |  | [default to undefined]
**default_value** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**key** | **string** |  | [default to undefined]
**overridable** | **boolean** |  | [optional] [default to undefined]
**required** | **boolean** |  | [optional] [default to undefined]
**template_string** | **string** |  | [optional] [default to undefined]
**type** | [**ParameterType**](ParameterType.md) |  | [default to undefined]
**value_type** | [**ValueDataType**](ValueDataType.md) |  | [optional] [default to undefined]

## Example

```typescript
import { CreateParameterConfigReq } from 'rcabench-client';

const instance: CreateParameterConfigReq = {
    category,
    default_value,
    description,
    key,
    overridable,
    required,
    template_string,
    type,
    value_type,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
