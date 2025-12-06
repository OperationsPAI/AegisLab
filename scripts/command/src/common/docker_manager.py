import time
from pathlib import Path

import yaml
from pydantic import BaseModel, Field
from python_on_whales import DockerClient

from src.common.common import console

__all__ = ["DockerManager"]


class DockerSessionData(BaseModel):
    """Session data for Docker manager."""

    client: DockerClient | None = None
    compose_file: Path | None = None
    services_running: bool = False
    running_services: list[str] = Field(default_factory=list)

    model_config = {"arbitrary_types_allowed": True}


class DockerManager:
    """Docker Compose manager with singleton pattern per compose file.

    Usage:
    with DockerManager(compose_file=Path("docker-compose.yaml")) as docker_manager:
        client = docker_manager.get_client()

    To explicitly stop services:
        docker_manager.stop_services()
    """

    _instances: dict[Path, "DockerManager"] = {}
    _sessions: dict[Path, DockerSessionData] = {}

    def __new__(
        cls,
        compose_file: Path | None = None,
        max_retries: int = 60,
        startup_wait: int = 10,
        services_list: list[str] | None = None,
    ):
        """Create or return existing singleton instance for the given compose file."""
        if compose_file is None:
            instance = super().__new__(cls)
            instance._is_singleton = False
            return instance

        # Resolve to absolute path for consistent key
        resolved_path = compose_file.resolve()

        if resolved_path not in cls._instances:
            instance = super().__new__(cls)
            cls._instances[resolved_path] = instance
            instance._initialized = False
            instance._is_singleton = True

        return cls._instances[resolved_path]

    def __init__(
        self,
        compose_file: Path | None = None,
        max_retries: int = 60,
        startup_wait: int = 10,
        services_list: list[str] | None = None,
    ) -> None:
        """Initialize Docker Compose manager.

        Args:
            compose_file: Path to docker-compose file
            max_retries: Maximum number of retries when waiting for services
            startup_wait: Additional wait time after services are ready (seconds)
            services_list: List of service names to monitor
        """
        if hasattr(self, "_initialized") and self._initialized:
            return

        self.compose_file = compose_file.resolve() if compose_file else None
        self.max_retries = max_retries
        self.startup_wait = startup_wait
        self.services_list = services_list
        self._client: DockerClient | None = None

        if not hasattr(self, "_is_singleton"):
            self._is_singleton = False
            if compose_file:
                self._initialize()
        else:
            self._initialized = True

    def __enter__(self) -> "DockerManager":
        """Context manager entry: ensure services are started."""
        if not self._is_singleton or self.compose_file is None:
            return self

        # Check if there is already a valid session
        if self.compose_file not in self._sessions or not self._is_session_valid():
            self._initialize_session()

        # Load session data
        session_data = self._sessions[self.compose_file]
        self._client = session_data.client

        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit: maintain singleton state, do not stop services."""
        pass

    def _is_session_valid(self) -> bool:
        """Check if the current session is valid."""
        if self.compose_file is None:
            return False

        session_data = self._sessions.get(self.compose_file)
        if not session_data or not session_data.services_running:
            return False

        # Verify services are still running
        try:
            if session_data.client and self.container_names:
                containers = session_data.client.compose.ps()
                running_containers = [
                    c.name for c in containers if c.state.status == "running"
                ]
                all_running = all(
                    name in running_containers for name in self.container_names
                )
                return all_running

        except Exception:
            return False

        return False

    def _initialize_session(self):
        """Initialize a new session and start services."""
        if self.compose_file is None:
            raise ValueError("compose_file must be provided for session initialization")

        self._initialize()

        # Store session information
        self._sessions[self.compose_file] = DockerSessionData(
            client=self._client,
            compose_file=self.compose_file,
            services_running=self._start_services(),
            running_services=self._get_running_service_names(),
        )

    def _initialize(self):
        """Initialize Docker client and load compose file."""
        if self.compose_file is None:
            return

        self._client = DockerClient(compose_files=[self.compose_file])

        with open(self.compose_file, encoding="utf-8") as f:
            compose_data = yaml.safe_load(f)

        services_data = compose_data.get("services", {})
        if services_data:
            if not self.services_list:
                self.services_list = list(compose_data.get("services", {}).keys())

            self.container_names: list[str] = []
            for value in services_data.values():
                if "container_name" in value:
                    self.container_names.append(value["container_name"])

    def _start_services(self) -> bool:
        """Start Docker Compose services and wait for them to be ready."""
        if self._client is None:
            raise RuntimeError("Docker client not initialized")

        console.print("\n🧹 Cleaning up existing containers...")
        try:
            self._client.compose.down(volumes=True, remove_orphans=True)
        except Exception:
            console.print("⚠️  No existing containers to clean up.")

        console.print("\n🚀 Starting Docker Compose services...")
        try:
            self._client.compose.up(
                detach=True, build=True, services=self.services_list
            )
        except Exception as e:
            console.print(f"⚠️  Some services may have failed to start: {e}")
            return False

        # Wait for services to be ready
        console.print("⏳ Waiting for services to be ready...")
        retry_count = 0

        while retry_count < self.max_retries:
            try:
                if self.container_names:
                    containers = self._client.compose.ps()
                    running_containers = [
                        c.name for c in containers if c.state.status == "running"
                    ]
                    all_running = all(
                        name in running_containers for name in self.container_names
                    )

                    if all_running:
                        console.print(
                            "\n✅ All specified services are running successfully"
                        )
                        time.sleep(self.startup_wait)
                        return True

            except Exception as e:
                console.print(f"\n⏳ Services not ready yet: {e}")

            time.sleep(2)
            retry_count += 1

        # Failed to start services
        self._print_failure_diagnostics()
        self._client.compose.down(volumes=True)
        raise RuntimeError("Docker Compose services failed to start")

    def _print_failure_diagnostics(self):
        """Print diagnostic information when services fail to start."""
        if self._client is None:
            return

        console.print("\n❌ Services failed to start in time")
        console.print("\n📋 Service status:")
        try:
            services = self._client.compose.ps()
            for service in services:
                console.print(f"  - {service.name}: {service.state.status}")
        except Exception:
            pass

        console.print("\n📋 Logs:")
        try:
            self._client.compose.logs(tail="50")
        except Exception:
            pass

    def _get_running_service_names(self) -> list[str]:
        """Get list of running service names."""
        if self._client is None:
            return []

        try:
            services = self._client.compose.ps()
            return [s.name for s in services if s.state.status == "running"]
        except Exception:
            return []

    def get_client(self) -> DockerClient:
        """Get the Docker client."""
        if self._client is None:
            raise RuntimeError("Docker client not initialized")
        return self._client

    def stop_services(self, remove_volumes: bool = True):
        """Explicitly stop Docker Compose services."""
        if self._client is None:
            return

        console.print("\n🛑 Stopping Docker Compose services...")
        try:
            self._client.compose.down(volumes=remove_volumes)
            console.print("✅ Docker Compose services stopped")

            # Update session state
            if self.compose_file and self.compose_file in self._sessions:
                self._sessions[self.compose_file].services_running = False
        except Exception as e:
            console.print(f"[bold red]Error stopping services: {e}[/bold red]")

    def is_running(self) -> bool:
        """Check if services are currently running."""
        return self._is_session_valid()

    def get_service_status(self) -> dict[str, str]:
        """Get status of all services."""
        if self._client is None:
            return {}

        try:
            services = self._client.compose.ps()
            return {s.name: s.state.status or "unknown" for s in services}
        except Exception:
            return {}

    @classmethod
    def clear_sessions(cls):
        """Clear all cached sessions and instances."""
        cls._sessions.clear()
        cls._instances.clear()
