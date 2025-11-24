# ContainerRef


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** |  | 
**version** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.container_ref import ContainerRef

# TODO update the JSON string below
json = "{}"
# create an instance of ContainerRef from a JSON string
container_ref_instance = ContainerRef.from_json(json)
# print the JSON string representation of the object
print(ContainerRef.to_json())

# convert the object into a dict
container_ref_dict = container_ref_instance.to_dict()
# create an instance of ContainerRef from a dict
container_ref_from_dict = ContainerRef.from_dict(container_ref_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


