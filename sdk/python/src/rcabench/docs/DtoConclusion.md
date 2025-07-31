# DtoConclusion


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**level** | **str** | 例如 service level | [optional] 
**metric** | **str** | 例如 topk | [optional] 
**rate** | **float** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_conclusion import DtoConclusion

# TODO update the JSON string below
json = "{}"
# create an instance of DtoConclusion from a JSON string
dto_conclusion_instance = DtoConclusion.from_json(json)
# print the JSON string representation of the object
print(DtoConclusion.to_json())

# convert the object into a dict
dto_conclusion_dict = dto_conclusion_instance.to_dict()
# create an instance of DtoConclusion from a dict
dto_conclusion_from_dict = DtoConclusion.from_dict(dto_conclusion_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


