import os
from dataclasses import dataclass

from pydantic import StrictStr

from rcabench.openapi.api.authentication_api import AuthenticationApi
from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.configuration import Configuration
from rcabench.openapi.models.login_req import LoginReq


@dataclass(kw_only=True)
class SessionData:
    access_token: StrictStr | None = None
    api_client: ApiClient | None = None


class RCABenchClient:
    """
    RCABench client supporting both username/password and token-based authentication.

    - Token-based auth (for K8s jobs):
        client = RCABenchClient(base_url="...", token="...")
        or via environment variable RCABENCH_TOKEN

    - Username/password auth (for interactive use):
        client = RCABenchClient(base_url="...", username="...", password="...")
        or via environment variables RCABENCH_USERNAME, RCABENCH_PASSWORD
    """

    _instances: dict[tuple[str, str, str | None], "RCABenchClient"] = {}
    _sessions: dict[tuple[str, str, str | None], SessionData] = {}

    def __new__(
        cls,
        base_url: str | None = None,
        username: str | None = None,
        password: str | None = None,
        token: str | None = None,
    ):
        # Parse actual configuration values
        actual_base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        actual_token = token or os.getenv("RCABENCH_TOKEN")
        actual_username = username or os.getenv("RCABENCH_USERNAME")
        actual_password = password or os.getenv("RCABENCH_PASSWORD")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"

        # Token auth takes precedence over username/password
        if actual_token:
            instance_key = (actual_base_url, actual_token, None)
        else:
            assert actual_username is not None, "username or RCABENCH_USERNAME is not set (or use token/RCABENCH_TOKEN)"
            assert actual_password is not None, "password or RCABENCH_PASSWORD is not set (or use token/RCABENCH_TOKEN)"
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
        token: str | None = None,
    ):
        # Avoid duplicate initialization of the same instance
        if hasattr(self, "_initialized") and self._initialized:
            return

        self.base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        self.token = token or os.getenv("RCABENCH_TOKEN")
        self.username = username or os.getenv("RCABENCH_USERNAME")
        self.password = password or os.getenv("RCABENCH_PASSWORD")

        assert self.base_url is not None, "base_url or RCABENCH_BASE_URL is not set"

        # Token auth takes precedence
        if self.token:
            self.instance_key = (self.base_url, self.token, None)
        else:
            assert self.username is not None, "username or RCABENCH_USERNAME is not set (or use token/RCABENCH_TOKEN)"
            assert self.password is not None, "password or RCABENCH_PASSWORD is not set (or use token/RCABENCH_TOKEN)"
            self.instance_key = (self.base_url, self.username, self.password)

        self._initialized = True

    def __enter__(self):
        # Check if there is already a valid session
        if self.instance_key not in self._sessions or not self._is_session_valid():
            self._authenticate()
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

    def _authenticate(self) -> None:
        """Authenticate using either token or username/password"""
        if self.token:
            # Direct token authentication (for K8s jobs using service tokens)
            self._sessions[self.instance_key] = SessionData(
                access_token=self.token,
                api_client=None,
            )
        else:
            # Username/password login
            self._login()

    def _login(self) -> None:
        """Login using username and password"""
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
            self._authenticate()

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
