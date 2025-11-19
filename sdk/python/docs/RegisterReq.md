# RegisterReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email** | **str** |  | 
**password** | **str** |  | 
**username** | **str** |  | 

## Example

```python
from openapi.models.register_req import RegisterReq

# TODO update the JSON string below
json = "{}"
# create an instance of RegisterReq from a JSON string
register_req_instance = RegisterReq.from_json(json)
# print the JSON string representation of the object
print(RegisterReq.to_json())

# convert the object into a dict
register_req_dict = register_req_instance.to_dict()
# create an instance of RegisterReq from a dict
register_req_from_dict = RegisterReq.from_dict(register_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


