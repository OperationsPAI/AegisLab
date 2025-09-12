# DtoMetricValue


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**timestamp** | **str** |  | [optional] 
**unit** | **str** |  | [optional] 
**value** | **float** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_metric_value import DtoMetricValue

# TODO update the JSON string below
json = "{}"
# create an instance of DtoMetricValue from a JSON string
dto_metric_value_instance = DtoMetricValue.from_json(json)
# print the JSON string representation of the object
print(DtoMetricValue.to_json())

# convert the object into a dict
dto_metric_value_dict = dto_metric_value_instance.to_dict()
# create an instance of DtoMetricValue from a dict
dto_metric_value_from_dict = DtoMetricValue.from_dict(dto_metric_value_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


