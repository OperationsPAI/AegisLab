# DtoRawDataReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithms** | **List[str]** |  | [optional] 
**datasets** | **List[str]** |  | [optional] 
**execution_ids** | **List[int]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_raw_data_req import DtoRawDataReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRawDataReq from a JSON string
dto_raw_data_req_instance = DtoRawDataReq.from_json(json)
# print the JSON string representation of the object
print(DtoRawDataReq.to_json())

# convert the object into a dict
dto_raw_data_req_dict = dto_raw_data_req_instance.to_dict()
# create an instance of DtoRawDataReq from a dict
dto_raw_data_req_from_dict = DtoRawDataReq.from_dict(dto_raw_data_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


