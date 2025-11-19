# ChaosResources


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**app_labels** | **List[str]** |  | [optional] 
**container_names** | **List[str]** |  | [optional] 
**database_app_names** | **List[str]** |  | [optional] 
**dns_app_names** | **List[str]** |  | [optional] 
**http_app_names** | **List[str]** |  | [optional] 
**jvm_app_names** | **List[str]** |  | [optional] 
**network_pairs** | [**List[ChaosPair]**](ChaosPair.md) |  | [optional] 

## Example

```python
from openapi.models.chaos_resources import ChaosResources

# TODO update the JSON string below
json = "{}"
# create an instance of ChaosResources from a JSON string
chaos_resources_instance = ChaosResources.from_json(json)
# print the JSON string representation of the object
print(ChaosResources.to_json())

# convert the object into a dict
chaos_resources_dict = chaos_resources_instance.to_dict()
# create an instance of ChaosResources from a dict
chaos_resources_from_dict = ChaosResources.from_dict(chaos_resources_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


