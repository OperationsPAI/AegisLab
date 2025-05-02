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
    def wrapper(*args, **kwargs) -> Union[Dict, Response, ModelHTTPError]:
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
        max_retries: int = 3,
        backoff_factor: float = 0.5,
    ):
        self.base_url = base_url
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
        headers: Optional[Dict] = None,
        params: Optional[Dict] = None,
        json: Optional[Any] = None,
        timeout: Optional[float] = None,
        stream: bool = False,
    ) -> Response:
        full_url = f"{self.base_url}{endpoint}"

        for attempt in range(self.max_retries):
            try:
                response = self.session.request(
                    method=method,
                    url=full_url,
                    headers=headers,
                    params=params,
                    json=json,
                    timeout=timeout,
                    stream=stream,
                )
                response.raise_for_status()
                return response

            except HTTPError as e:
                status_code = e.response.status_code if e.response is not None else 500
                if 500 <= status_code < 600 and attempt < self.max_retries:
                    self._handle_retry(attempt, e)
                    continue

                error_message = f"Server returned {status_code}"
                if e.response:
                    try:
                        error_data = e.response.json()
                        error_message = (
                            error_data.get("message")
                            or error_data.get("detail")
                            or error_message
                        )
                    except (ValueError, AttributeError):
                        error_message = e.response.text[:200]

                raise HTTPClientError(
                    message=error_message,
                    status_code=status_code,
                    url=full_url,
                ) from e

            except (Timeout, NewConnectionError, RequestException) as e:
                if attempt == self.max_retries - 1:
                    raise HTTPClientError(
                        message=f"Connection failed after {self.max_retries} retries: {str(e)}",
                        status_code=503,
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
    def delete(
        self,
        endpoint: str,
        params: Optional[Dict] = None,
    ) -> Response:
        return self._request("DELETE", endpoint, params=params)

    @handle_http_errors
    def get(
        self,
        endpoint: str,
        headers: Optional[Dict] = None,
        params: Optional[Dict] = None,
        stream: bool = False,
        timeout: Optional[float] = None,
    ) -> Response:
        return self._request(
            "GET",
            endpoint,
            headers=headers,
            params=params,
            stream=stream,
            timeout=timeout,
        )

    @handle_http_errors
    def post(
        self,
        endpoint: str,
        json: Dict,
    ) -> Response:
        return self._request("POST", endpoint, json=json)

    @handle_http_errors
    def put(
        self,
        endpoint: str,
        json: Dict,
    ) -> Response:
        return self._request("PUT", endpoint, json=json)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.session.close()
