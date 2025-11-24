# CreateParameterConfigReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | [**ParameterCategory**](ParameterCategory.md) |  | 
**default_value** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**key** | **str** |  | 
**required** | **bool** |  | [optional] 
**template_string** | **str** |  | [optional] 
**type** | [**ParameterType**](ParameterType.md) |  | 

## Example

```python
from rcabench.openapi.models.create_parameter_config_req import CreateParameterConfigReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateParameterConfigReq from a JSON string
create_parameter_config_req_instance = CreateParameterConfigReq.from_json(json)
# print the JSON string representation of the object
print(CreateParameterConfigReq.to_json())

# convert the object into a dict
create_parameter_config_req_dict = create_parameter_config_req_instance.to_dict()
# create an instance of CreateParameterConfigReq from a dict
create_parameter_config_req_from_dict = CreateParameterConfigReq.from_dict(create_parameter_config_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


