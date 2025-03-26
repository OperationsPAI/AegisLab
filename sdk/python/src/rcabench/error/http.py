from typing import Optional


class HttpClientError(Exception):
    """自定义HTTP客户端异常"""

    def __init__(
        self, message: str, status_code: Optional[int] = None, url: Optional[str] = None
    ):
        self.message = message
        self.status_code = status_code
        self.url = url
        super().__init__(f"HTTP Error ({status_code}) at {url}: {message}")
