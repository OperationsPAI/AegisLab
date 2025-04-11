from typing import Any, Callable, Dict, Optional, Union
from ..error.http import HTTPClientError
from ..logger import logger
from ..model.error import ModelHTTPError
from functools import wraps
from requests.adapters import HTTPAdapter
from requests.exceptions import HTTPError, RequestException, Timeout
from requests import Response
from urllib3.exceptions import NewConnectionError
import inspect
import requests
import time

__all__ = ["HTTPClient"]


def handle_http_errors(func: Callable):
    @wraps(func)
    def wrapper(*args, **kwargs) -> Union[Any, ModelHTTPError]:
        try:
            resp = func(*args, **kwargs)
            sig = inspect.signature(func)
            bound_args = sig.bind(*args, **kwargs)
            if "stream" in bound_args.arguments:
                return resp

            resp_data = resp.json()
            return resp_data.get("data")

        except HTTPClientError as e:
            # 统一记录日志并返回错误响应
            logger.error(f"API request failed: {e.url} -> {e.message}")
            return ModelHTTPError(
                status_code=e.status_code,
                detail=e.message,
                path=args[1],
                method=str.upper(func.__name__),
            )

        except Exception as e:
            logger.error(f"Unknown error: {str(e)}")
            return ModelHTTPError(
                status_code=500,
                detail=str(e),
                path=args[1],
                method=str.upper(func.__name__),
            )

    return wrapper


class HTTPClient:
    def __init__(
        self,
        base_url: str,
        timeout: int = 10,
        max_retries: int = 3,
        backoff_factor: float = 0.5,
    ):
        self.base_url = base_url
        self.timeout = timeout
        self.max_retries = max_retries
        self.backoff_factor = backoff_factor

        # 配置Session对象复用TCP连接
        self.session = requests.Session()
        adapter = HTTPAdapter(max_retries=max_retries)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    def _request(
        self,
        method: str,
        endpoint: str,
        params: Optional[Dict] = None,
        json: Optional[Any] = None,
        stream: bool = False,
        retries: int = 3,
    ) -> Response:
        full_url = f"{self.base_url}{endpoint}"

        for attempt in range(retries):
            try:
                response = self.session.request(
                    method=method,
                    url=full_url,
                    params=params,
                    json=json,
                    timeout=self.timeout,
                    stream=stream,
                )
                response.raise_for_status()
                return response

            except HTTPError as e:
                status_code = e.response.status_code if e.response else 500
                if 500 <= status_code < 600 and attempt < self.max_retries:
                    self._handle_retry(attempt, e)
                    continue

                error_message = f"Server returned {status_code}"
                if e.response and e.response.content:
                    try:
                        error_data = e.response.json()
                        error_message = (
                            error_data.get("message")
                            or error_data.get("detail")
                            or error_message
                        )
                    except (ValueError, AttributeError):
                        pass

                raise HTTPClientError(
                    message=error_message,
                    status_code=status_code,
                    url=full_url,
                ) from e

            except (Timeout, NewConnectionError, RequestException) as e:
                if attempt == self.max_retries - 1:
                    raise HTTPClientError(
                        message=f"Connection failed after {self.max_retries} retries: {str(e)}",
                        status_code=503,  # 服务不可用状态码
                        url=full_url,
                    ) from e
                self._handle_retry(attempt, e)

    def _handle_retry(self, attempt: int, error: Exception) -> None:
        sleep_time = self.backoff_factor * (2**attempt)
        logger.warning(
            f"Attempt {attempt + 1} failed: {error}. Retrying in {sleep_time:.1f}s..."
        )
        time.sleep(sleep_time)

    @handle_http_errors
    def delete(self, endpoint: str, params: Optional[Dict] = None) -> Any:
        return self._request("DELETE", endpoint, params=params)

    @handle_http_errors
    def get(
        self, endpoint: str, params: Optional[Dict] = None, stream: bool = False
    ) -> Any:
        return self._request("GET", endpoint, params=params, stream=stream)

    @handle_http_errors
    def post(self, endpoint: str, payload: Dict) -> Any:
        return self._request("POST", endpoint, json=payload)

    @handle_http_errors
    def put(self, endpoint: str, payload: Dict) -> Any:
        return self._request("PUT", endpoint, json=payload)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.session.close()
