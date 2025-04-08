from typing import Dict, List, Optional
from .stream import Stream

__all__ = ["Task"]


class Task:
    URL_PREFIX = "/tasks"

    URL_ENDPOINTS = {
        "get_detail": "/{task_id}",
        "get_stream": "/{task_id}/stream",
    }

    def __init__(self, client, max_connections: int = 10):
        self.client = client
        self.stream = Stream(client.base_url, max_connections)

    async def get_stream(
        self,
        task_ids: List[str],
        timeout: Optional[float] = None,
    ) -> Dict[str, any]:
        urls = [
            f"{self.URL_PREFIX}{self.URL_ENDPOINTS['get_stream'].format(task_id=task_id)}"
            for task_id in task_ids
        ]

        report = await self.stream.start_multiple_stream(
            urls, client_ids=task_ids, timeout=timeout
        )
        return report
