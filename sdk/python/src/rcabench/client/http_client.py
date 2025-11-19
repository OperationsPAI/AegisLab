import os
from dataclasses import dataclass
from typing import TYPE_CHECKING

from pydantic import StrictStr

from rcabench.openapi.api.authentication_api import AuthenticationApi
from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.configuration import Configuration
from rcabench.openapi.models.login_req import LoginReq

if TYPE_CHECKING:
    from rcabench.client.http_client import RCABenchClient


@dataclass(kw_only=True)
class SessionData:
    access_token: StrictStr | None = None
    api_client: ApiClient | None = None


class RCABenchClient:
    """
    Usage:
    with RCABenchClient() as api_client:
        container_api = rcabench.openapi.ContainersApi(api_client)
        containers = container_api.api_v2_containers_get()
        print(f"Containers: {containers.data}")
    """

    _instances: dict[tuple[str, str, str], "RCABenchClient"] = {}
    _sessions: dict[tuple[str, str, str], SessionData] = {}

    def __new__(
        cls,
        base_url: str | None = None,
        username: str | None = None,
        password: str | None = None,
    ):
        # Parse actual configuration values
        actual_base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        actual_username = username or os.getenv("RCABENCH_USERNAME")
        actual_password = password or os.getenv("RCABENCH_PASSWORD")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"
        assert actual_username is not None, "username or RCABENCH_USERNAME is not set"
        assert actual_password is not None, "password or RCABENCH_PASSWORD is not set"

        # Use (base_url, username) as unique identifier
        instance_key = (actual_base_url, actual_username, actual_password)

        if instance_key not in cls._instances:
            instance = super().__new__(cls)
            cls._instances[instance_key] = instance
            instance._initialized = False

        return cls._instances[instance_key]

    def __init__(
        self,
        base_url: str | None = None,
        username: str | None = None,
        password: str | None = None,
    ):
        # Avoid duplicate initialization of the same instance
        if hasattr(self, "_initialized") and self._initialized:
            return

        self.base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        self.username = username or os.getenv("RCABENCH_USERNAME")
        self.password = password or os.getenv("RCABENCH_PASSWORD")

        assert self.username is not None, "username or RCABENCH_USERNAME is not set"
        assert self.password is not None, "password or RCABENCH_PASSWORD is not set"
        assert self.base_url is not None, "base_url or RCABENCH_BASE_URL is not set"

        self.instance_key = (self.base_url, self.username, self.password)
        self._initialized = True

    def __enter__(self):
        # Check if there is already a valid session
        if self.instance_key not in self._sessions or not self._is_session_valid():
            self._login()
        return self._get_authenticated_client()

    def __exit__(self, exc_type, exc_val, exc_tb):
        # Do not close session, maintain singleton state
        pass

    def _is_session_valid(self) -> bool:
        """Check if the current session is valid"""
        session_data = self._sessions.get(self.instance_key)
        if not session_data:
            return False

        # More complex session validity checks can be added here, such as checking if token is expired
        # Currently simply check if access_token exists
        return session_data.access_token is not None

    def _login(self) -> None:
        config = Configuration(host=self.base_url)
        with ApiClient(config) as api_client:
            auth_api = AuthenticationApi(api_client)
            assert self.base_url is not None
            assert self.username is not None
            assert self.password is not None
            login_request = LoginReq(username=self.username, password=self.password)
            response = auth_api.login(request=login_request)
            assert response.data is not None

            # Store session information in class-level cache
            self._sessions[self.instance_key] = SessionData(
                access_token=response.data.token,
                api_client=None,  # Will be created on demand
            )

    def _get_authenticated_client(self) -> ApiClient:
        if self.instance_key not in self._sessions or not self._is_session_valid():
            self._login()

        session_data = self._sessions[self.instance_key]

        # If api_client has not been created or needs to be updated, create a new one
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
    def clear_sessions(cls):
        cls._sessions.clear()
        cls._instances.clear()
