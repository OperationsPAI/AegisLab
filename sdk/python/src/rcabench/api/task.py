from ..client.http_client import HTTPClient

__all__ = ["Task"]


class Task:
    URL_PREFIX = "/tasks"

    URL_ENDPOINTS = {
        "get_detail": "/{task_id}",
    }

    def __init__(self, client: HTTPClient, api_version: str):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"
