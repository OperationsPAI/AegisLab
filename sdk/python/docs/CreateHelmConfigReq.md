# CreateHelmConfigReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**chart_name** | **str** |  | 
**ns_prefix** | **str** |  | 
**repo_name** | **str** |  | 
**repo_url** | **str** |  | 
**values** | **object** |  | [optional] 

## Example

```python
from rcabench.openapi.models.create_helm_config_req import CreateHelmConfigReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateHelmConfigReq from a JSON string
create_helm_config_req_instance = CreateHelmConfigReq.from_json(json)
# print the JSON string representation of the object
print(CreateHelmConfigReq.to_json())

# convert the object into a dict
create_helm_config_req_dict = create_helm_config_req_instance.to_dict()
# create an instance of CreateHelmConfigReq from a dict
create_helm_config_req_from_dict = CreateHelmConfigReq.from_dict(create_helm_config_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


