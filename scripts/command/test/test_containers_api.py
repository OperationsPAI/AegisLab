import pytest
from conftest import DEFAULT_PAGE, DEFAULT_PAGE_SIZE

from rcabench.openapi import ApiClient
from rcabench.openapi.api.containers_api import ContainersApi
from rcabench.openapi.models.container_type import ContainerType
from rcabench.openapi.models.create_container_req import CreateContainerReq
from rcabench.openapi.models.create_container_version_req import (
    CreateContainerVersionReq,
)
from rcabench.openapi.models.create_helm_config_req import CreateHelmConfigReq
from rcabench.openapi.models.create_parameter_config_req import (
    CreateParameterConfigReq,
)
from rcabench.openapi.models.page_size import PageSize
from rcabench.openapi.models.parameter_category import ParameterCategory
from rcabench.openapi.models.parameter_type import ParameterType
from rcabench.openapi.models.status_type import StatusType


@pytest.fixture(scope="module")
def containers_api(rcabench_client: ApiClient) -> ContainersApi:
    return ContainersApi(rcabench_client)


class TestContainers:
    @pytest.mark.order(1)
    @pytest.mark.parametrize(
        "container_id,expected_name",
        [
            (1, "detector"),  # Container 1
            (2, "traceback"),  # Container 2
            (3, "clickhouse"),  # Container 3
            (4, "ts"),  # Container 4
            (5, "ts_cn"),  # Container 5
        ],
    )
    def test_get_container_by_id(
        self, containers_api: ContainersApi, container_id: int, expected_name: str
    ) -> None:
        """Test retrieving containers by ID using initial data.

        Verifies that each container ID maps to the expected container name from data.json.
        """
        resp = containers_api.get_container_by_id(container_id=container_id)
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected container data in response"
        assert resp.data.id == container_id, "Container ID mismatch"
        assert resp.data.name == expected_name, (
            f"Expected container name '{expected_name}', got '{resp.data.name}'"
        )
        assert resp.data.type is not None, "Expected container type"

    @pytest.mark.order(2)
    @pytest.mark.parametrize(
        "page,size,expected_length",
        [
            (
                1,
                PageSize.Small,
                5,
            ),  # Page 1: all 5 containers (detector, traceback, clickhouse, ts, ts_cn)
            (2, PageSize.Small, 0),  # Page 2: no more containers
            (1, PageSize.Medium, 5),  # Page 1 with medium size: all 5 containers
            (1, PageSize.Large, 5),  # Page 1 with large size: all 5 containers
        ],
    )
    def test_list_containers_with_pagination(
        self,
        containers_api: ContainersApi,
        page: int,
        size: PageSize,
        expected_length: int,
    ) -> None:
        """Test listing containers with pagination parameters.

        Initial data has 5 containers (IDs 1-5):
        - Container 1: detector (Algorithm)
        - Container 2: traceback (Algorithm)
        - Container 3: clickhouse (Benchmark)
        - Container 4: ts (Pedestal)
        - Container 5: ts_cn (Pedestal)
        """
        resp = containers_api.list_containers(page=page, size=size)
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected data in response"
        assert resp.data.items is not None, "Expected items in data"
        assert isinstance(resp.data.items, list), "Expected items to be a list"
        assert len(resp.data.items) == expected_length, (
            f"Expected exactly {expected_length} items on page {page}"
        )

        if expected_length > 0:
            # Verify pagination metadata
            assert resp.data.pagination is not None, "Expected pagination info"
            assert resp.data.pagination.total is not None, "Expected total count"
            assert resp.data.pagination.total == 5, (
                "Expected total of 5 containers in initial data"
            )
            assert resp.data.pagination.page == page, "Page number mismatch"
            assert resp.data.pagination.size == size, "Page size mismatch"

    @pytest.mark.order(3)
    @pytest.mark.parametrize(
        "container_type,is_public,status,expected_length",
        [
            (ContainerType.Algorithm, None, None, 2),
            (None, True, None, 5),
            (None, None, StatusType.Enabled, 5),
        ],
    )
    def test_list_containers_with_filters(
        self,
        containers_api: ContainersApi,
        container_type: ContainerType,
        is_public: bool | None,
        status: StatusType | None,
        expected_length: int,
    ) -> None:
        """Test listing containers with type and status filters."""
        resp = containers_api.list_containers(
            type=container_type,
            is_public=is_public,
            status=status,
        )
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected data in response"
        assert resp.data.items is not None, "Expected items in data"
        assert isinstance(resp.data.items, list), "Expected items to be a list"
        assert len(resp.data.items) == expected_length, (
            f"Expected exactly {expected_length} items"
        )

        if expected_length > 0:
            # Verify pagination metadata
            assert resp.data.pagination is not None, "Expected pagination info"
            assert resp.data.pagination.total is not None, "Expected total count"
            assert resp.data.pagination.total == expected_length, (
                f"Expected total of {expected_length} containers in initial data"
            )
            assert resp.data.pagination.page == DEFAULT_PAGE, "Page number mismatch"
            assert resp.data.pagination.size == DEFAULT_PAGE_SIZE, "Page size mismatch"

    @pytest.mark.order(4)
    def test_create_container(self, containers_api: ContainersApi) -> None:
        """Test creating a new container."""
        create_req = CreateContainerReq(
            name="test-container",
            type=ContainerType.Algorithm,
            readme="Test container created by SDK test",
            is_public=False,
            version=CreateContainerVersionReq(
                name="1.0.0",
                image_ref="docker.io/test/image:latest",
                command="python app.py",
                env_vars=[
                    CreateParameterConfigReq(
                        key="ENV",
                        type=ParameterType.Fixed,
                        category=ParameterCategory.EnvVars,
                    )
                ],
                github_link="test/repo",
                helm_config=CreateHelmConfigReq(
                    repo_name="test-repo",
                    chart_name="test-chart",
                    repo_url="https://example.com/charts",
                    ns_prefix="ts",
                ),
            ),
        )
        resp = containers_api.create_container(request=create_req)
        assert resp.code == 201, f"Expected 201 Created, got {resp.code}"
        assert resp.data is not None, "Expected container data in response"
        assert resp.data.id is not None, "Expected container ID"
        assert resp.data.name == create_req.name, "Container name mismatch"


class TestContainerVersions:
    @pytest.mark.order(5)
    @pytest.mark.parametrize(
        "container_id,version_id",
        [
            (1, 1),  # Container 1, Version 1
            (2, 2),  # Container 2, Version 2
            (3, 3),  # Container 3, Version 3
            (4, 4),  # Container 4, Version 4
            (5, 5),  # Container 5, Version 5
        ],
    )
    def test_get_container_version_by_id(
        self,
        containers_api: ContainersApi,
        container_id: int,
        version_id: int,
    ) -> None:
        """Test retrieving a container version by ID using initial data.

        Assumes version IDs are auto-incremented starting from 1, matching container creation order.
        """
        resp = containers_api.get_container_version_by_id(
            container_id=container_id, version_id=version_id
        )
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected version data in response"
        assert resp.data.id == version_id, (
            f"Version ID mismatch: expected {version_id}, got {resp.data.id}"
        )
        assert resp.data.name is not None, "Expected version name"
        assert resp.data.image_ref is not None, "Expected image reference"

    @pytest.mark.order(6)
    @pytest.mark.parametrize(
        "container_id,page,size,expected_length",
        [
            (1, 1, PageSize.Small, 1),  # Container 1 (detector) has at least 1 version
            (2, 1, PageSize.Small, 1),  # Container 2 (traceback) has at least 1 version
            (
                3,
                1,
                PageSize.Small,
                1,
            ),  # Container 3 (clickhouse) has at least 1 version
            (4, 1, PageSize.Small, 1),  # Container 4 (ts) has at least 1 version
            (5, 1, PageSize.Small, 1),  # Container 5 (ts_cn) has at least 1 version
        ],
    )
    def test_list_container_versions_with_pagination(
        self,
        containers_api: ContainersApi,
        container_id: int,
        page: int,
        size: PageSize,
        expected_length: int,
    ) -> None:
        """Test listing container versions with pagination.

        Each container in initial data has at least 1 version.
        """
        resp = containers_api.list_container_versions(
            container_id=container_id, page=page, size=size
        )
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected data in response"
        assert resp.data.items is not None, "Expected items in data"
        assert isinstance(resp.data.items, list), "Expected items to be a list"
        assert len(resp.data.items) == expected_length, (
            f"Expected exactly {expected_length} items on page {page}"
        )

        # Verify pagination metadata
        if expected_length > 0:
            # Verify pagination metadata
            assert resp.data.pagination is not None, "Expected pagination info"
            assert resp.data.pagination.total is not None, "Expected total count"
            assert resp.data.pagination.total == 1, (
                "Expected total of 1 container in initial data"
            )
            assert resp.data.pagination.page == page, "Page number mismatch"
            assert resp.data.pagination.size == size, "Page size mismatch"

    @pytest.mark.order(7)
    @pytest.mark.parametrize(
        "container_id,status,expected_length",
        [
            (1, StatusType.Enabled, 1),  # Filter enabled versions for container 1
            (2, StatusType.Enabled, 1),  # Filter enabled versions for container 2
        ],
    )
    def test_list_container_versions_with_status_filter(
        self,
        containers_api: ContainersApi,
        container_id: int,
        status: StatusType | None,
        expected_length: int,
    ) -> None:
        """Test listing container versions with status filter."""
        resp = containers_api.list_container_versions(
            container_id=container_id, status=status
        )
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected data in response"
        assert resp.data.items is not None, "Expected items in data"
        assert isinstance(resp.data.items, list), "Expected items to be a list"
        assert len(resp.data.items) == expected_length, (
            f"Expected exactly {expected_length} items"
        )

        if expected_length > 0:
            # Verify pagination metadata
            assert resp.data.pagination is not None, "Expected pagination info"
            assert resp.data.pagination.total is not None, "Expected total count"
            assert resp.data.pagination.total == expected_length, (
                f"Expected total of {expected_length} containers in initial data"
            )
            assert resp.data.pagination.page == DEFAULT_PAGE, "Page number mismatch"
            assert resp.data.pagination.size == DEFAULT_PAGE_SIZE, "Page size mismatch"

    @pytest.mark.order(8)
    def test_create_container_version(self, containers_api: ContainersApi) -> None:
        """Test creating a new container version."""
        create_container_version_req = CreateContainerVersionReq(
            name="1.0.1",
            image_ref="docker.io/test/image:latest",
            command="python app.py",
            env_vars=[
                CreateParameterConfigReq(
                    key="ENV",
                    type=ParameterType.Fixed,
                    category=ParameterCategory.EnvVars,
                )
            ],
            github_link="test/repo",
            helm_config=CreateHelmConfigReq(
                repo_name="test-repo",
                chart_name="test-chart",
                repo_url="https://example.com/charts",
                ns_prefix="ts",
            ),
        )
        resp = containers_api.create_container_version(
            container_id=1, request=create_container_version_req
        )
        assert resp.code == 201, f"Expected 201 Created, got {resp.code}"
        assert resp.data is not None, "Expected version data in response"
        assert resp.data.id is not None, "Expected version ID"
