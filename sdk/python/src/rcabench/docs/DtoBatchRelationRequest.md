# DtoBatchRelationRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**operations** | [**List[DtoRelationOperation]**](DtoRelationOperation.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_batch_relation_request import DtoBatchRelationRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoBatchRelationRequest from a JSON string
dto_batch_relation_request_instance = DtoBatchRelationRequest.from_json(json)
# print the JSON string representation of the object
print(DtoBatchRelationRequest.to_json())

# convert the object into a dict
dto_batch_relation_request_dict = dto_batch_relation_request_instance.to_dict()
# create an instance of DtoBatchRelationRequest from a dict
dto_batch_relation_request_from_dict = DtoBatchRelationRequest.from_dict(dto_batch_relation_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


