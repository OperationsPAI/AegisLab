# DtoDatasetV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**checksum** | **str** | File checksum | [optional] 
**created_at** | **str** | Creation time | [optional] 
**data_source** | **str** | Data source description | [optional] 
**description** | **str** | Dataset description | [optional] 
**download_url** | **str** | Download URL | [optional] 
**file_count** | **int** | File count | [optional] 
**format** | **str** | Data format | [optional] 
**id** | **int** | Unique identifier | [optional] 
**injections** | [**List[DtoDatasetV2InjectionRelationResponse]**](DtoDatasetV2InjectionRelationResponse.md) | Associated fault injections | [optional] 
**is_public** | **bool** | Whether public | [optional] 
**labels** | [**List[DatabaseLabel]**](DatabaseLabel.md) | Associated labels | [optional] 
**name** | **str** | Dataset name | [optional] 
**project** | [**DatabaseProject**](DatabaseProject.md) | Associated project info | [optional] 
**project_id** | **int** | Project ID | [optional] 
**status** | **int** | Status | [optional] 
**type** | **str** | Dataset type | [optional] 
**updated_at** | **str** | Update time | [optional] 
**version** | **str** | Dataset version | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_response import DtoDatasetV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2Response from a JSON string
dto_dataset_v2_response_instance = DtoDatasetV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2Response.to_json())

# convert the object into a dict
dto_dataset_v2_response_dict = dto_dataset_v2_response_instance.to_dict()
# create an instance of DtoDatasetV2Response from a dict
dto_dataset_v2_response_from_dict = DtoDatasetV2Response.from_dict(dto_dataset_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


