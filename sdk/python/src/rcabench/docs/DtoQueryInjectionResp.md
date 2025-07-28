# DtoQueryInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** | 基准数据库，添加索引 | [optional] 
**created_at** | **str** | 创建时间，添加时间索引 | [optional] 
**description** | **str** | 描述（可选字段） | [optional] 
**display_config** | **str** | 面向用户的展示配置 | [optional] 
**end_time** | **str** | 预计故障结束时间，添加时间索引 | [optional] 
**engine_config** | **str** | 面向系统的运行配置 | [optional] 
**fault_type** | **int** | 故障类型，添加复合索引 | [optional] 
**ground_truth** | [**HandlerGroundtruth**](HandlerGroundtruth.md) |  | [optional] 
**id** | **int** | 唯一标识 | [optional] 
**injection_name** | **str** | 在k8s资源里注入的名字 | [optional] 
**pre_duration** | **int** | 正常数据时间 | [optional] 
**start_time** | **str** | 预计故障开始时间，添加时间索引 | [optional] 
**status** | **int** | 状态，添加复合索引 | [optional] 
**task** | [**DatabaseTask**](DatabaseTask.md) | 外键关联 | [optional] 
**task_id** | **str** | 从属什么 taskid，添加复合索引 | [optional] 
**updated_at** | **str** | 更新时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_query_injection_resp import DtoQueryInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoQueryInjectionResp from a JSON string
dto_query_injection_resp_instance = DtoQueryInjectionResp.from_json(json)
# print the JSON string representation of the object
print(DtoQueryInjectionResp.to_json())

# convert the object into a dict
dto_query_injection_resp_dict = dto_query_injection_resp_instance.to_dict()
# create an instance of DtoQueryInjectionResp from a dict
dto_query_injection_resp_from_dict = DtoQueryInjectionResp.from_dict(dto_query_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


