import os
import secrets
import time
from dataclasses import dataclass
from hashlib import sha256
from hmac import new as hmac_new

from pydantic import StrictStr

from rcabench.openapi.api.authentication_api import AuthenticationApi
from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.configuration import Configuration


@dataclass(kw_only=True)
class SessionData:
    access_token: StrictStr | None = None
    api_client: ApiClient | None = None


class RCABenchClient:
    """
    RCABench client supporting access-key and token-based authentication.

    - Token-based auth (for K8s jobs):
        client = RCABenchClient(base_url="...", token="...")
        or via environment variable RCABENCH_TOKEN

    - Access-key auth (recommended for SDK use):
        client = RCABenchClient(base_url="...", access_key="...", secret_key="...")
        or via environment variables RCABENCH_ACCESS_KEY, RCABENCH_SECRET_KEY
    """

    _instances: dict[tuple[str, str, str | None], "RCABenchClient"] = {}
    _sessions: dict[tuple[str, str, str | None], SessionData] = {}
    _token_exchange_path = "/api/v2/auth/access-key/token"

    def __new__(
        cls,
        base_url: str | None = None,
        access_key: str | None = None,
        secret_key: str | None = None,
        token: str | None = None,
    ):
        # Parse actual configuration values
        actual_base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        actual_token = token or os.getenv("RCABENCH_TOKEN")
        actual_access_key = access_key or os.getenv("RCABENCH_ACCESS_KEY")
        actual_secret_key = secret_key or os.getenv("RCABENCH_SECRET_KEY")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"

        # Token auth takes precedence over access-key authentication
        if actual_token:
            instance_key = (actual_base_url, actual_token, None)
        else:
            assert actual_access_key is not None, (
                "access_key or RCABENCH_ACCESS_KEY is not set (or use token/RCABENCH_TOKEN)"
            )
            assert actual_secret_key is not None, (
                "secret_key or RCABENCH_SECRET_KEY is not set (or use token/RCABENCH_TOKEN)"
            )
            instance_key = (actual_base_url, actual_access_key, actual_secret_key)

        if instance_key not in cls._instances:
            instance = super().__new__(cls)
            cls._instances[instance_key] = instance
            instance._initialized = False

        return cls._instances[instance_key]

    def __init__(
        self,
        base_url: str | None = None,
        access_key: str | None = None,
        secret_key: str | None = None,
        token: str | None = None,
    ):
        # Avoid duplicate initialization of the same instance
        if hasattr(self, "_initialized") and self._initialized:
            return

        self.base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        self.token = token or os.getenv("RCABENCH_TOKEN")
        self.access_key = access_key or os.getenv("RCABENCH_ACCESS_KEY")
        self.secret_key = secret_key or os.getenv("RCABENCH_SECRET_KEY")

        assert self.base_url is not None, "base_url or RCABENCH_BASE_URL is not set"

        # Token auth takes precedence
        if self.token:
            self.instance_key = (self.base_url, self.token, None)
        else:
            assert self.access_key is not None, (
                "access_key or RCABENCH_ACCESS_KEY is not set (or use token/RCABENCH_TOKEN)"
            )
            assert self.secret_key is not None, (
                "secret_key or RCABENCH_SECRET_KEY is not set (or use token/RCABENCH_TOKEN)"
            )
            self.instance_key = (self.base_url, self.access_key, self.secret_key)

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
        """Authenticate using either token or access-key credentials."""
        if self.token:
            # Direct token authentication.
            self._sessions[self.instance_key] = SessionData(
                access_token=self.token,
                api_client=None,
            )
        else:
            self._exchange_access_key_token()

    def _exchange_access_key_token(self) -> None:
        """Exchange access_key/secret_key for a bearer token."""
        config = Configuration(host=self.base_url)
        with ApiClient(config) as api_client:
            auth_api = AuthenticationApi(api_client)
            assert self.base_url is not None
            assert self.access_key is not None
            assert self.secret_key is not None
            timestamp = str(int(time.time()))
            nonce = secrets.token_hex(16)
            signature = self._sign_access_key_request(
                secret_key=self.secret_key,
                method="POST",
                path=self._token_exchange_path,
                access_key=self.access_key,
                timestamp=timestamp,
                nonce=nonce,
            )
            response = auth_api.exchange_access_key_token(
                x_access_key=self.access_key,
                x_timestamp=timestamp,
                x_nonce=nonce,
                x_signature=signature,
            )
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

    @staticmethod
    def _sign_access_key_request(
        secret_key: str,
        method: str,
        path: str,
        access_key: str,
        timestamp: str,
        nonce: str,
    ) -> str:
        canonical = "\n".join([method.upper(), path, access_key, timestamp, nonce])
        return hmac_new(
            secret_key.encode("utf-8"),
            canonical.encode("utf-8"),
            sha256,
        ).hexdigest()

    @classmethod
    def clear_sessions(cls):
        cls._sessions.clear()
        cls._instances.clear()
