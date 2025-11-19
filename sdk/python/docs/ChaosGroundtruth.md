# ChaosGroundtruth


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**container** | **List[str]** |  | [optional] 
**function** | **List[str]** |  | [optional] 
**metric** | **List[str]** |  | [optional] 
**pod** | **List[str]** |  | [optional] 
**service** | **List[str]** |  | [optional] 
**span** | **List[str]** |  | [optional] 

## Example

```python
from openapi.models.chaos_groundtruth import ChaosGroundtruth

# TODO update the JSON string below
json = "{}"
# create an instance of ChaosGroundtruth from a JSON string
chaos_groundtruth_instance = ChaosGroundtruth.from_json(json)
# print the JSON string representation of the object
print(ChaosGroundtruth.to_json())

# convert the object into a dict
chaos_groundtruth_dict = chaos_groundtruth_instance.to_dict()
# create an instance of ChaosGroundtruth from a dict
chaos_groundtruth_from_dict = ChaosGroundtruth.from_dict(chaos_groundtruth_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


