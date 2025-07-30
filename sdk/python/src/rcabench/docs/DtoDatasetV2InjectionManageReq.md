# DtoDatasetV2InjectionManageReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**add_injections** | **List[int]** | List of fault injection IDs to add | [optional] 
**remove_injections** | **List[int]** | List of fault injection IDs to remove | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_injection_manage_req import DtoDatasetV2InjectionManageReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2InjectionManageReq from a JSON string
dto_dataset_v2_injection_manage_req_instance = DtoDatasetV2InjectionManageReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2InjectionManageReq.to_json())

# convert the object into a dict
dto_dataset_v2_injection_manage_req_dict = dto_dataset_v2_injection_manage_req_instance.to_dict()
# create an instance of DtoDatasetV2InjectionManageReq from a dict
dto_dataset_v2_injection_manage_req_from_dict = DtoDatasetV2InjectionManageReq.from_dict(dto_dataset_v2_injection_manage_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


