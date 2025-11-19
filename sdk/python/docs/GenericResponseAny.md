# GenericResponseAny


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | **object** | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_any import GenericResponseAny

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseAny from a JSON string
generic_response_any_instance = GenericResponseAny.from_json(json)
# print the JSON string representation of the object
print(GenericResponseAny.to_json())

# convert the object into a dict
generic_response_any_dict = generic_response_any_instance.to_dict()
# create an instance of GenericResponseAny from a dict
generic_response_any_from_dict = GenericResponseAny.from_dict(generic_response_any_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


