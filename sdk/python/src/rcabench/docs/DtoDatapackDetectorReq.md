# DtoDatapackDetectorReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**datapacks** | **List[str]** |  | 
**tag** | **str** | Tag filter for filtering execution results | [optional] 

## Example

```python
from rcabench.openapi.models.dto_datapack_detector_req import DtoDatapackDetectorReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatapackDetectorReq from a JSON string
dto_datapack_detector_req_instance = DtoDatapackDetectorReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatapackDetectorReq.to_json())

# convert the object into a dict
dto_datapack_detector_req_dict = dto_datapack_detector_req_instance.to_dict()
# create an instance of DtoDatapackDetectorReq from a dict
dto_datapack_detector_req_from_dict = DtoDatapackDetectorReq.from_dict(dto_datapack_detector_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


