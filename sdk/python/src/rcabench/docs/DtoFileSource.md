# DtoFileSource

File source configuration for uploads

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**file_name** | **str** | 通过 multipart/form-data 上传的文件会自动处理 这里只是为了文档说明 @Description Filename of the uploaded file | [optional] 
**size** | **int** | @Description Size of the uploaded file in bytes | [optional] 

## Example

```python
from rcabench.openapi.models.dto_file_source import DtoFileSource

# TODO update the JSON string below
json = "{}"
# create an instance of DtoFileSource from a JSON string
dto_file_source_instance = DtoFileSource.from_json(json)
# print the JSON string representation of the object
print(DtoFileSource.to_json())

# convert the object into a dict
dto_file_source_dict = dto_file_source_instance.to_dict()
# create an instance of DtoFileSource from a dict
dto_file_source_from_dict = DtoFileSource.from_dict(dto_file_source_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


