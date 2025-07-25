# DtoListInjectionsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoInjectionItem]**](DtoInjectionItem.md) |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_list_injections_resp import DtoListInjectionsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoListInjectionsResp from a JSON string
dto_list_injections_resp_instance = DtoListInjectionsResp.from_json(json)
# print the JSON string representation of the object
print(DtoListInjectionsResp.to_json())

# convert the object into a dict
dto_list_injections_resp_dict = dto_list_injections_resp_instance.to_dict()
# create an instance of DtoListInjectionsResp from a dict
dto_list_injections_resp_from_dict = DtoListInjectionsResp.from_dict(dto_list_injections_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


