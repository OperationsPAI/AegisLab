import pytest
from conftest import DEFAULT_PAGE, DEFAULT_PAGE_SIZE
from rcabench.openapi import ApiClient
from rcabench.openapi.api.projects_api import ProjectsApi
from rcabench.openapi.models.create_project_req import CreateProjectReq
from rcabench.openapi.models.page_size import PageSize
from rcabench.openapi.models.status_type import StatusType


@pytest.fixture(scope="module")
def projects_api(rcabench_client: ApiClient) -> ProjectsApi:
    return ProjectsApi(rcabench_client)


class TestProjects:
    @pytest.mark.order(1)
    @pytest.mark.parametrize(
        "project_id,expected_name",
        [
            (1, "pair_diagnosis"),  # Project 1
        ],
    )
    def test_get_project_by_id(
        self, projects_api: ProjectsApi, project_id: int, expected_name: str
    ) -> None:
        """Test retrieving projects by ID using initial data.

        Verifies that each project ID maps to the expected project name from data.json.
        """
        resp = projects_api.get_project_by_id(project_id=project_id)
        assert resp.code == 200, f"Expected 200 OK, got {resp.code}"
        assert resp.data is not None, "Expected project data in response"
        assert resp.data.id == project_id, "Project ID mismatch"
        assert resp.data.name == expected_name, (
            f"Expected project name '{expected_name}', got '{resp.data.name}'"
        )
        assert resp.data.description is not None, "Expected project description"

    @pytest.mark.order(2)
    @pytest.mark.parametrize(
        "page,size,expected_length",
        [
            (1, PageSize.Small, 1),  # Page 1: 1 project (pair_diagnosis)
            (2, PageSize.Small, 0),  # Page 2: no more projects
            (1, PageSize.Medium, 1),  # Page 1 with medium size: 1 project
            (1, PageSize.Large, 1),  # Page 1 with large size: 1 project
        ],
    )
    def test_list_projects_with_pagination(
        self, projects_api: ProjectsApi, page: int, size: PageSize, expected_length: int
    ) -> None:
        """Test listing projects with pagination parameters.

        Initial data has 1 project (ID 1):
        - Project 1: pair_diagnosis
        """
        resp = projects_api.list_projects(page=page, size=size)
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
                "Expected total of 1 project in initial data"
            )
            assert resp.data.pagination.page == page, "Page number mismatch"
            assert resp.data.pagination.size == size, "Page size mismatch"

    @pytest.mark.order(3)
    @pytest.mark.parametrize(
        "is_public,status,expected_length",
        [
            (None, StatusType.Enabled, 1),  # Filter by enabled status
            (False, None, 1),  # No public projects in initial data
        ],
    )
    def test_list_projects_with_filters(
        self,
        projects_api: ProjectsApi,
        is_public: bool | None,
        status: StatusType | None,
        expected_length: int,
    ) -> None:
        """Test listing projects with status filters."""
        resp = projects_api.list_projects(
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
                f"Expected total of {expected_length} projects in initial data"
            )
            assert resp.data.pagination.page == DEFAULT_PAGE, "Page number mismatch"
            assert resp.data.pagination.size == DEFAULT_PAGE_SIZE, "Page size mismatch"

    @pytest.mark.order(4)
    def test_create_project(self, projects_api: ProjectsApi) -> None:
        """Test creating a new project."""
        create_req = CreateProjectReq(
            name="test-project",
            description="Test project created by SDK test",
        )
        resp = projects_api.create_project(request=create_req)
        assert resp.code == 201, f"Expected 201 Created, got {resp.code}"
        assert resp.data is not None, "Expected project data in response"
        assert resp.data.id is not None, "Expected project ID"
        assert resp.data.name == create_req.name, "Project name mismatch"
