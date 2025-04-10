from typing import Any, Callable, Dict, Optional, Tuple, Type, TypeVar, Union
from ..model.error import (
    FieldValidationIssue,
    ModelHTTPError,
    ModelValidationError,
)
from functools import wraps
from pydantic import BaseModel, ValidationError as PydanticValidationError
import inspect

__all__ = ["validate_request_response"]

T = TypeVar("T", bound=BaseModel)


class ValidationHelper:
    """验证逻辑公共工具类"""

    @staticmethod
    def safe_validate(
        model: Type[T], input_data: Union[Any]
    ) -> Tuple[Optional[T], Optional[ModelValidationError]]:
        """
        泛型安全校验函数

        Args:
            model: 继承自BaseModel的泛型模型类
            input_data: 待校验的原始数据字典

        Returns:
            - Tuple[validated_model, None] 验证成功
            - Tuple[None, ModelValidationError] 验证失败

        Example:
            >>> class UserModel(BaseModel):
            ...     name: str
            >>> data, error = safe_validate(UserModel, {"name": "Alice"})
            >>> print(isinstance(data, UserModel))  # True
        """
        try:
            validated = model.model_validate(input_data)
            return validated, None

        except PydanticValidationError as ex:
            issues = []
            for err in ex.errors():
                issues.append(
                    FieldValidationIssue(
                        field=".".join(map(str, err["loc"])),
                        message=err["msg"],
                        error_type=err["type"],
                    )
                )

            return None, ModelValidationError(
                detail=f"Invalid {model.__name__} payload",
                issues=issues,
            )

    @staticmethod
    def prepare_input_data(
        func: Callable, *args, **kwargs
    ) -> Tuple[Optional[Dict], Optional[ModelValidationError]]:
        """准备待验证的输入数据"""
        # 获取被装饰函数的参数名称
        sig = inspect.signature(func)
        try:
            bound_args = sig.bind(*args, **kwargs)
        except TypeError as e:
            return None, ModelValidationError(str(e))
        return {**bound_args.arguments}, None

    @staticmethod
    def validate_request(
        request_model: Type[T], input_data: dict, func: Callable
    ) -> Tuple[Optional[Dict], Optional[ModelValidationError]]:
        """请求参数验证（同步）"""
        is_instance_method = (
            len(input_data) > 0
            and hasattr(list(input_data.values())[0], func.__name__)
            and inspect.ismethod(getattr(list(input_data.values())[0], func.__name__))
        )

        if is_instance_method:
            instance = list(input_data.values())[0]
            validation_data = {
                k: v for k, v in input_data.items() if k != list(input_data.keys())[0]
            }
        else:
            validation_data = input_data

        req, err = ValidationHelper.safe_validate(request_model, validation_data)
        if err is not None:
            return None, err

        validated_data = req.model_dump(exclude_unset=True)
        if is_instance_method:
            validated_data.update({list(input_data.keys())[0]: instance})

        return validated_data, None

    @staticmethod
    def validate_response(
        response_model: Type[T], result: Any
    ) -> Tuple[Optional[Dict], Optional[ModelValidationError]]:
        """响应数据验证（同步）"""
        resp, err = ValidationHelper.safe_validate(response_model, result)
        if err is not None:
            return None, err

        return resp, None


def validate_request_response(
    request_model: Optional[type[T]] = None,
    response_model: Optional[type[T]] = None,
) -> Callable:
    """
    同步装饰器工厂，用于验证请求参数和响应数据

    Args:
        request_model: 请求参数的Pydantic模型
        response_model: 响应数据的Pydantic模型
    """

    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs) -> Union[T, ModelHTTPError, ModelValidationError]:
            # 构建输入数据（保留self/cls引用）
            input_data, err = ValidationHelper.prepare_input_data(func, *args, **kwargs)
            if err is not None:
                return err

            if request_model is not None:
                input_data, err = ValidationHelper.validate_request(
                    request_model, input_data, func
                )
                if err is not None:
                    return err

            result = func(**input_data)
            if isinstance(result, ModelHTTPError):
                return result

            # 验证响应数据
            if response_model is not None:
                result, err = ValidationHelper.validate_response(response_model, result)
                if err is not None:
                    return err

                return result

            return result

        return wrapper

    return decorator


def async_validate_request_response(
    request_model: Optional[type[T]] = None,
    response_model: Optional[type[T]] = None,
) -> Callable:
    """
    异步装饰器工厂，用于验证请求参数和响应数据

    Args:
        request_model: 请求参数的Pydantic模型
        response_model: 响应数据的Pydantic模型
    """

    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def async_wrapper(
            *args, **kwargs
        ) -> Union[T, ModelHTTPError, ModelValidationError]:
            # 构建输入数据（保留self/cls引用）
            input_data, err = ValidationHelper.prepare_input_data(func, *args, **kwargs)
            if err is not None:
                return err

            if request_model is not None:
                input_data, err = ValidationHelper.validate_request(
                    request_model, input_data, func
                )
                if err is not None:
                    return err

            result = await func(**input_data)
            if isinstance(result, ModelHTTPError):
                return result

            # 验证响应数据
            if response_model is not None:
                result, err = ValidationHelper.validate_response(response_model, result)
                if err is not None:
                    return err

                return result

            return result

        return async_wrapper

    return decorator
