# EvaluateDatapackSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | [**ContainerRef**](ContainerRef.md) |  | 
**datapack** | **str** |  | 
**filter_labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.evaluate_datapack_spec import EvaluateDatapackSpec

# TODO update the JSON string below
json = "{}"
# create an instance of EvaluateDatapackSpec from a JSON string
evaluate_datapack_spec_instance = EvaluateDatapackSpec.from_json(json)
# print the JSON string representation of the object
print(EvaluateDatapackSpec.to_json())

# convert the object into a dict
evaluate_datapack_spec_dict = evaluate_datapack_spec_instance.to_dict()
# create an instance of EvaluateDatapackSpec from a dict
evaluate_datapack_spec_from_dict = EvaluateDatapackSpec.from_dict(evaluate_datapack_spec_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


