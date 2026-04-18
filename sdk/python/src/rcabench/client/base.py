from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import ClassVar

from pydantic import StrictStr

from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.configuration import Configuration


@dataclass(kw_only=True)
class SessionData:
    access_token: StrictStr | None = None
    api_client: ApiClient | None = None


CacheKey = tuple[str, str, str | None]


class BaseRCABenchClient(ABC):
    """
    Shared authenticated client lifecycle for hand-written RCABench clients.

    Subclasses own:
    - auth input resolution
    - instance/session cache keys
    - _authenticate implementation
    """

    _instances: ClassVar[dict[CacheKey, "BaseRCABenchClient"]] = {}
    _sessions: ClassVar[dict[CacheKey, SessionData]] = {}
    base_url: str
    instance_key: CacheKey
    _initialized: bool

    def __enter__(self) -> ApiClient:
        if self.instance_key not in self.__class__._sessions or not self._is_session_valid():
            self._authenticate()
        return self._get_authenticated_client()

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        pass

    def _is_session_valid(self) -> bool:
        session_data = self.__class__._sessions.get(self.instance_key)
        if not session_data:
            return False
        return session_data.access_token is not None

    @abstractmethod
    def _authenticate(self) -> None:
        raise NotImplementedError

    def _get_authenticated_client(self) -> ApiClient:
        if self.instance_key not in self.__class__._sessions or not self._is_session_valid():
            self._authenticate()

        session_data = self.__class__._sessions[self.instance_key]
        bearer_token = session_data.access_token
        assert bearer_token is not None, "Access token is missing in session data"

        if not session_data.api_client:
            auth_config = Configuration(
                host=self.base_url,
                api_key={"BearerAuth": bearer_token},
                api_key_prefix={"BearerAuth": "Bearer"},
            )
            session_data.api_client = ApiClient(auth_config)

        return session_data.api_client

    def get_client(self) -> ApiClient:
        return self._get_authenticated_client()

    @classmethod
    def clear_sessions(cls) -> None:
        cls._sessions.clear()
        cls._instances.clear()
