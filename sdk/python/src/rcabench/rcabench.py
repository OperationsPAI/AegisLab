from .api import Algorithm, Dataset, Evaluation, Injection, Task
from .client.http_client import HttpClient


class RCABenchSDK:
    def __init__(self, base_url: str, max_connections: int = 10):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/") + "/api/v1"

        client = HttpClient(self.base_url)
        self.algorithm = Algorithm(client)
        self.dataset = Dataset(client)
        self.evaluation = Evaluation(client)
        self.injection = Injection(client)
        self.task = Task(client, max_connections)
