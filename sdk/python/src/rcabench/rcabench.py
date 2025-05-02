from .api import Algorithm, Dataset, Evaluation, Injection, Task, Trace
from .client.http_client import HTTPClient


class RCABenchSDK:
    def __init__(self, base_url: str, api_version: str = "/api/v1"):
        client = HTTPClient(base_url.rstrip("/"))

        self.algorithm = Algorithm(client, api_version)
        self.dataset = Dataset(client, api_version)
        self.evaluation = Evaluation(client, api_version)
        self.injection = Injection(client, api_version)
        self.task = Task(client, api_version)
        self.trace = Trace(client, api_version)
