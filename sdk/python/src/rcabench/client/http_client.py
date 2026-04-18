import os
import secrets
import time
from hashlib import sha256
from hmac import new as hmac_new
from typing import ClassVar

from rcabench.client.base import BaseRCABenchClient, CacheKey, SessionData
from rcabench.openapi.api.authentication_api import AuthenticationApi
from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.configuration import Configuration


class RCABenchClient(BaseRCABenchClient):
    """
    RCABench public client supporting API-key authentication.

    Auth credentials are loaded from environment variables only:
    - RCABENCH_BASE_URL or `base_url=...`
    - RCABENCH_KEY_ID
    - RCABENCH_KEY_SECRET
    """

    _instances: ClassVar[dict[CacheKey, BaseRCABenchClient]] = {}
    _sessions: ClassVar[dict[CacheKey, SessionData]] = {}
    _token_exchange_path = "/api/v2/auth/api-key/token"

    def __new__(
        cls,
        base_url: str | None = None,
    ):
        actual_base_url = base_url or os.getenv("RCABENCH_BASE_URL")
        actual_key_id = os.getenv("RCABENCH_KEY_ID")
        actual_key_secret = os.getenv("RCABENCH_KEY_SECRET")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"
        assert actual_key_id is not None, "RCABENCH_KEY_ID is not set"
        assert actual_key_secret is not None, "RCABENCH_KEY_SECRET is not set"
        instance_key = (actual_base_url, actual_key_id, actual_key_secret)

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
        actual_key_id = os.getenv("RCABENCH_KEY_ID")
        actual_key_secret = os.getenv("RCABENCH_KEY_SECRET")

        assert actual_base_url is not None, "base_url or RCABENCH_BASE_URL is not set"
        assert actual_key_id is not None, "RCABENCH_KEY_ID is not set"
        assert actual_key_secret is not None, "RCABENCH_KEY_SECRET is not set"
        self.base_url = actual_base_url
        self.key_id = actual_key_id
        self.key_secret = actual_key_secret
        self.instance_key = (self.base_url, self.key_id, self.key_secret)

        self._initialized = True

    def _authenticate(self) -> None:
        self._exchange_api_key_token()

    def _exchange_api_key_token(self) -> None:
        config = Configuration(host=self.base_url)
        with ApiClient(config) as api_client:
            auth_api = AuthenticationApi(api_client)
            assert self.base_url is not None
            assert self.key_id is not None
            assert self.key_secret is not None
            timestamp = str(int(time.time()))
            nonce = secrets.token_hex(16)
            signature = self._sign_api_key_request(
                key_secret=self.key_secret,
                method="POST",
                path=self._token_exchange_path,
                timestamp=timestamp,
                nonce=nonce,
            )
            response = auth_api.exchange_api_key_token(
                x_key_id=self.key_id,
                x_timestamp=timestamp,
                x_nonce=nonce,
                x_signature=signature,
            )
            assert response.data is not None
            self.__class__._sessions[self.instance_key] = SessionData(
                access_token=response.data.token,
                api_client=None,
            )

    @staticmethod
    def _sign_api_key_request(
        key_secret: str,
        method: str,
        path: str,
        timestamp: str,
        nonce: str,
    ) -> str:
        body_hash = sha256(b"").hexdigest()
        canonical = "\n".join([method.upper(), path, timestamp, nonce, body_hash])
        return hmac_new(
            key_secret.encode("utf-8"),
            canonical.encode("utf-8"),
            sha256,
        ).hexdigest()
