from .api import Trace
from .client.http_client import HTTPClient


class RCABenchSDK:
    def __init__(self, base_url: str, api_version: str = "/api/v1"):
        client = HTTPClient(base_url.rstrip("/"))
        self.trace = Trace(client, api_version)
