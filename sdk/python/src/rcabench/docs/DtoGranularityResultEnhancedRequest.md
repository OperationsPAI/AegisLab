# DtoGranularityResultEnhancedRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**datapack_id** | **int** | Required if no execution_id | [optional] 
**duration** | **int** | Execution duration in seconds | 
**results** | [**List[DtoGranularityResultItem]**](DtoGranularityResultItem.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_granularity_result_enhanced_request import DtoGranularityResultEnhancedRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGranularityResultEnhancedRequest from a JSON string
dto_granularity_result_enhanced_request_instance = DtoGranularityResultEnhancedRequest.from_json(json)
# print the JSON string representation of the object
print(DtoGranularityResultEnhancedRequest.to_json())

# convert the object into a dict
dto_granularity_result_enhanced_request_dict = dto_granularity_result_enhanced_request_instance.to_dict()
# create an instance of DtoGranularityResultEnhancedRequest from a dict
dto_granularity_result_enhanced_request_from_dict = DtoGranularityResultEnhancedRequest.from_dict(dto_granularity_result_enhanced_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


