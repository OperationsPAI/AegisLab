# DtoListInjectionsReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**batches** | **List[str]** |  | [optional] 
**benchmarks** | **List[str]** |  | [optional] 
**custom_end_time** | **str** |  | [optional] 
**custom_start_time** | **str** |  | [optional] 
**envs** | **List[str]** | Label filter parameters (optional) | [optional] 
**fault_types** | **List[int]** |  | [optional] 
**ids** | **List[int]** |  | [optional] 
**lookback** | **str** |  | [optional] 
**names** | **List[str]** | Basic filter parameters (optional) | [optional] 
**page_num** | **int** |  | 
**page_size** | **int** |  | 
**statuses** | **List[int]** |  | [optional] 
**task_ids** | **List[str]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_list_injections_req import DtoListInjectionsReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoListInjectionsReq from a JSON string
dto_list_injections_req_instance = DtoListInjectionsReq.from_json(json)
# print the JSON string representation of the object
print(DtoListInjectionsReq.to_json())

# convert the object into a dict
dto_list_injections_req_dict = dto_list_injections_req_instance.to_dict()
# create an instance of DtoListInjectionsReq from a dict
dto_list_injections_req_from_dict = DtoListInjectionsReq.from_dict(dto_list_injections_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


