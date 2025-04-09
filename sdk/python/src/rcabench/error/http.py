class HTTPClientError(Exception):
    """自定义HTTP客户端异常"""

    def __init__(self, message: str, status_code: int, url: str):
        self.message = message
        self.status_code = status_code
        self.url = url
        super().__init__(f"HTTP Error ({status_code}) at {url}: {message}")
