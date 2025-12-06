import pytest
from conftest import DEFAULT_PAGE, DEFAULT_PAGE_SIZE
from rcabench.openapi import ApiClient
from rcabench.openapi.api.datasets_api import DatasetsApi
from rcabench.openapi.models.create_dataset_req import CreateDatasetReq
from rcabench.openapi.models.create_dataset_version_req import CreateDatasetVersionReq
from rcabench.openapi.models.page_size import PageSize
from rcabench.openapi.models.status_type import StatusType


@pytest.fixture(scope="module")
def datasets_api(rcabench_client: ApiClient) -> DatasetsApi:
    return DatasetsApi(rcabench_client)


class TestDatasets:
    @pytest.mark.order(1)
    @pytest.mark.parametrize(
        "dataset_id,expected_name",
        [
            (1, "rca_pair_diagnosis_dataset"),  # Dataset 1
        ],
    )
    def test_get_dataset_by_id(
        self, datasets_api: DatasetsApi, dataset_id: int, expected_name: str
    ) -> None:
        """Test retrieving datasets by ID using initial data.

        Verifies that each dataset ID maps to the expected dataset name from data.json.
        """
        resp = datasets_api.get_dataset_by_id(dataset_id=dataset_id)
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected dataset data in response"
        assert resp.data.id == dataset_id, "Dataset ID mismatch"
        assert resp.data.name == expected_name, (
            f"Expected dataset name '{expected_name}', got '{resp.data.name}'"
        )
        assert resp.data.type is not None, "Expected dataset type"
        assert resp.data.description is not None, "Expected dataset description"

    @pytest.mark.order(2)
    @pytest.mark.parametrize(
        "page,size,expected_length",
        [
            (1, PageSize.Small, 1),  # Page 1: 1 dataset (rca_pair_diagnosis_dataset)
            (2, PageSize.Small, 0),  # Page 2: no more datasets
            (1, PageSize.Medium, 1),  # Page 1 with medium size: 1 dataset
            (1, PageSize.Large, 1),  # Page 1 with large size: 1 dataset
        ],
    )
    def test_list_datasets_with_pagination(
        self, datasets_api: DatasetsApi, page: int, size: PageSize, expected_length: int
    ) -> None:
        """Test listing datasets with pagination parameters.

        Initial data has 1 dataset (ID 1):
        - Dataset 1: rca_pair_diagnosis_dataset (network type)
        """
        resp = datasets_api.list_datasets(page=page, size=size)
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
            assert resp.data.pagination.total == 1, (
                "Expected total of 1 dataset in initial data"
            )
            assert resp.data.pagination.page == page, "Page number mismatch"
            assert resp.data.pagination.size == size, "Page size mismatch"

    @pytest.mark.order(3)
    @pytest.mark.parametrize(
        "dataset_type,is_public,status,expected_length",
        [
            ("network", True, None, 1),  # Filter by network type
            ("network", None, StatusType.Enabled, 1),  # Filter by enabled status
        ],
    )
    def test_list_datasets_with_filters(
        self,
        datasets_api: DatasetsApi,
        dataset_type: str,
        is_public: bool | None,
        status: StatusType | None,
        expected_length: int,
    ) -> None:
        """Test listing datasets with type and status filters."""
        resp = datasets_api.list_datasets(
            type=dataset_type,
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
                f"Expected total of {expected_length} datasets in initial data"
            )
            assert resp.data.pagination.page == DEFAULT_PAGE, "Page number mismatch"
            assert resp.data.pagination.size == DEFAULT_PAGE_SIZE, "Page size mismatch"

    @pytest.mark.order(4)
    def test_create_dataset(self, datasets_api: DatasetsApi) -> None:
        """Test creating a new dataset."""
        create_req = CreateDatasetReq(
            name="test-dataset",
            description="Test dataset created by SDK test",
            type="network",
            is_public=False,
            version=CreateDatasetVersionReq(
                name="1.0.0",
            ),
        )
        resp = datasets_api.create_dataset(request=create_req)
        assert resp.code == 201, f"Expected 201 Created, got {resp.code}"
        assert resp.data is not None, "Expected dataset data in response"
        assert resp.data.id is not None, "Expected dataset ID"
        assert resp.data.name == create_req.name, "Dataset name mismatch"


class TestDatasetVersions:
    @pytest.mark.order(5)
    @pytest.mark.parametrize(
        "dataset_id,version_id",
        [
            (1, 1),  # Dataset 1, Version 1
        ],
    )
    def test_get_dataset_version_by_id(
        self,
        datasets_api: DatasetsApi,
        dataset_id: int,
        version_id: int,
    ) -> None:
        """Test retrieving a dataset version by ID using initial data.

        Assumes version IDs are auto-incremented starting from 1, matching dataset creation order.
        """
        resp = datasets_api.get_dataset_version_by_id(
            dataset_id=dataset_id, version_id=version_id
        )
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected version data in response"
        assert resp.data.id == version_id, (
            f"Version ID mismatch: expected {version_id}, got {resp.data.id}"
        )
        assert resp.data.name is not None, "Expected version name"

    @pytest.mark.order(6)
    @pytest.mark.parametrize(
        "dataset_id,page,size,expected_length",
        [
            (
                1,
                1,
                PageSize.Small,
                1,
            ),  # Dataset 1 (rca_pair_diagnosis_dataset) has at least 1 version
        ],
    )
    def test_list_dataset_versions_with_pagination(
        self,
        datasets_api: DatasetsApi,
        dataset_id: int,
        page: int,
        size: PageSize,
        expected_length: int,
    ) -> None:
        """Test listing dataset versions with pagination.

        Each dataset in initial data has at least 1 version.
        """
        resp = datasets_api.list_dataset_versions(
            dataset_id=dataset_id, page=page, size=size
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
                "Expected total of 1 dataset version in initial data"
            )
            assert resp.data.pagination.page == page, "Page number mismatch"
            assert resp.data.pagination.size == size, "Page size mismatch"

    @pytest.mark.order(7)
    @pytest.mark.parametrize(
        "dataset_id,status,expected_length",
        [
            (1, StatusType.Enabled, 1),  # Filter enabled versions for dataset 1
        ],
    )
    def test_list_dataset_versions_with_status_filter(
        self,
        datasets_api: DatasetsApi,
        dataset_id: int,
        status: StatusType | None,
        expected_length: int,
    ) -> None:
        """Test listing dataset versions with status filter."""
        resp = datasets_api.list_dataset_versions(dataset_id=dataset_id, status=status)
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
                f"Expected total of {expected_length} dataset versions in initial data"
            )
            assert resp.data.pagination.page == DEFAULT_PAGE, "Page number mismatch"
            assert resp.data.pagination.size == DEFAULT_PAGE_SIZE, "Page size mismatch"

    @pytest.mark.order(8)
    def test_create_dataset_version(self, datasets_api: DatasetsApi) -> None:
        """Test creating a new dataset version."""
        create_dataset_version_req = CreateDatasetVersionReq(
            name="1.0.1",
        )
        resp = datasets_api.create_dataset_version(
            dataset_id=1, request=create_dataset_version_req
        )
        assert resp.code == 201, f"Expected 201 Created, got {resp.code}"
        assert resp.data is not None, "Expected version data in response"
        assert resp.data.id is not None, "Expected version ID"
