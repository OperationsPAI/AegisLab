# DtoPermissionSearchRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**actions** | **List[str]** | Action filter | [optional] 
**created_at** | [**DtoDateRange**](DtoDateRange.md) | Common filters shortcuts | [optional] 
**description_pattern** | **str** | Fuzzy match for description | [optional] 
**display_name_pattern** | **str** | Fuzzy match for display name | [optional] 
**exclude_fields** | **List[str]** |  | [optional] 
**filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) | Filters | [optional] 
**include** | **List[str]** | Include related entities | [optional] 
**include_fields** | **List[str]** | Include/Exclude fields | [optional] 
**is_active** | **bool** |  | [optional] 
**is_system** | **bool** | Is system permission | [optional] 
**keyword** | **str** | Search keyword (for general text search) | [optional] 
**name_pattern** | **str** | Permission-specific filter shortcuts | [optional] 
**page** | **int** | Pagination | [optional] 
**project_id** | **int** |  | [optional] 
**resource_ids** | **List[int]** | Resource ID filter | [optional] 
**resource_names** | **List[str]** | Resource name filter | [optional] 
**role_ids** | **List[int]** | Role IDs that have this permission | [optional] 
**size** | **int** |  | [optional] 
**sort** | [**List[DtoSortOption]**](DtoSortOption.md) | Sort | [optional] 
**status** | **List[int]** |  | [optional] 
**updated_at** | [**DtoDateRange**](DtoDateRange.md) |  | [optional] 
**user_id** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_permission_search_request import DtoPermissionSearchRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPermissionSearchRequest from a JSON string
dto_permission_search_request_instance = DtoPermissionSearchRequest.from_json(json)
# print the JSON string representation of the object
print(DtoPermissionSearchRequest.to_json())

# convert the object into a dict
dto_permission_search_request_dict = dto_permission_search_request_instance.to_dict()
# create an instance of DtoPermissionSearchRequest from a dict
dto_permission_search_request_from_dict = DtoPermissionSearchRequest.from_dict(dto_permission_search_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


