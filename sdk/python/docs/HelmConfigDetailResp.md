# HelmConfigDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**full_chart** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**ns_prefix** | **str** |  | [optional] 
**port_template** | **str** |  | [optional] 
**repo_url** | **str** |  | [optional] 
**values** | **Dict[str, object]** |  | [optional] 

## Example

```python
from openapi.models.helm_config_detail_resp import HelmConfigDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of HelmConfigDetailResp from a JSON string
helm_config_detail_resp_instance = HelmConfigDetailResp.from_json(json)
# print the JSON string representation of the object
print(HelmConfigDetailResp.to_json())

# convert the object into a dict
helm_config_detail_resp_dict = helm_config_detail_resp_instance.to_dict()
# create an instance of HelmConfigDetailResp from a dict
helm_config_detail_resp_from_dict = HelmConfigDetailResp.from_dict(helm_config_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


