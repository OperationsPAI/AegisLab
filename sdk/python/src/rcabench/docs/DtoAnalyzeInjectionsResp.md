# DtoAnalyzeInjectionsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**efficiency** | **str** |  | [optional] 
**stats** | [**Dict[str, DtoInjectionStats]**](DtoInjectionStats.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_analyze_injections_resp import DtoAnalyzeInjectionsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAnalyzeInjectionsResp from a JSON string
dto_analyze_injections_resp_instance = DtoAnalyzeInjectionsResp.from_json(json)
# print the JSON string representation of the object
print(DtoAnalyzeInjectionsResp.to_json())

# convert the object into a dict
dto_analyze_injections_resp_dict = dto_analyze_injections_resp_instance.to_dict()
# create an instance of DtoAnalyzeInjectionsResp from a dict
dto_analyze_injections_resp_from_dict = DtoAnalyzeInjectionsResp.from_dict(dto_analyze_injections_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


