from typing import Any, Dict, Optional
from ..error.http import HttpClientError
from ..logger import logger
from functools import wraps
from requests.adapters import HTTPAdapter
from requests.exceptions import HTTPError, RequestException, Timeout
from requests import Response
import requests
import time

__all__ = ["HttpClient"]


def handle_http_errors(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        try:
            resp = func(*args, **kwargs)
            if "stream" in kwargs:
                return resp

            resp_data = resp.json()
            return resp_data.get("data", None)

        except HttpClientError as e:
            # 统一记录日志或返回错误响应
            logger.error(f"API request failed: {e.url} -> {e.message}")
            return None

        except Exception as e:
            logger.error(f"Unknown error: {str(e)}")
            return None

    return wrapper


class HttpClient:
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
        url: str,
        params: Optional[Dict] = None,
        json: Optional[Any] = None,
        stream: bool = False,
        retries: int = 3,
    ) -> Optional[Response]:
        full_url = f"{self.base_url}{url}"

        for attempt in range(retries):
            try:
                logger.info(f"Sending {method} request to {full_url}")

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
                status_code = e.response.status_code
                if 500 <= status_code < 600 and attempt < self.max_retries:
                    self._handle_retry(attempt, e)
                    continue
                raise HttpClientError(
                    message=f"Server returned {status_code}",
                    status_code=status_code,
                    url=url,
                ) from e
            except (Timeout, RequestException) as e:
                if attempt == self.max_retries:
                    raise HttpClientError(
                        message=f"Request failed after {self.max_retries} retries: {str(e)}",
                        url=url,
                    ) from e
                self._handle_retry(attempt, e)

        return None

    def _handle_retry(self, attempt: int, error: Exception):
        sleep_time = self.backoff_factor * (2**attempt)
        logger.warning(
            f"Attempt {attempt + 1} failed: {error}. Retrying in {sleep_time:.1f}s..."
        )
        time.sleep(sleep_time)

    @handle_http_errors
    def delete(self, url: str, params: Optional[Dict] = None) -> Any:
        return self._request("DELETE", url, params=params)

    @handle_http_errors
    def get(self, url: str, params: Optional[Dict] = None, stream: bool = False) -> Any:
        return self._request("GET", url, params=params, stream=stream)

    @handle_http_errors
    def post(self, url: str, payload: Dict) -> Any:
        return self._request("POST", url, json=payload)

    @handle_http_errors
    def put(self, url: str, payload: Dict) -> Any:
        return self._request("PUT", url, json=payload)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.session.close()
