import os
from typing import ClassVar

from rcabench.client.base import BaseRCABenchClient, CacheKey, SessionData


class RCABenchRuntimeClient(BaseRCABenchClient):
    """
    Runtime-only client for managed workloads.

    Auth credentials are loaded from environment variables only:
    - RCABENCH_BASE_URL or `base_url=...`
    - RCABENCH_SERVICE_TOKEN
    """

    _instances: ClassVar[dict[CacheKey, BaseRCABenchClient]] = {}
    _sessions: ClassVar[dict[CacheKey, SessionData]] = {}

    def __new__(
        cls,
        base_url: str | None = None,
    ):
        actual_base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        actual_service_token = os.getenv("RCABENCH_SERVICE_TOKEN")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"
        assert actual_service_token is not None, "RCABENCH_SERVICE_TOKEN is not set"

        instance_key = (actual_base_url, actual_service_token, None)

        if instance_key not in cls._instances:
            instance = super().__new__(cls)
            cls._instances[instance_key] = instance
            instance._initialized = False

        return cls._instances[instance_key]

    def __init__(
        self,
        base_url: str | None = None,
    ):
        if hasattr(self, "_initialized") and self._initialized:
            return

        actual_base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        actual_service_token = os.getenv("RCABENCH_SERVICE_TOKEN")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"
        assert actual_service_token is not None, "RCABENCH_SERVICE_TOKEN is not set"

        self.base_url = actual_base_url
        self.service_token = actual_service_token
        self.instance_key = (self.base_url, self.service_token, None)
        self._initialized = True

    def _authenticate(self) -> None:
        self.__class__._sessions[self.instance_key] = SessionData(
            access_token=self.service_token,
            api_client=None,
        )
