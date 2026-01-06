from typing import Any, ClassVar

import pytest
from rcabench.openapi import ApiClient
from rcabench.openapi.api.evaluations_api import EvaluationsApi
from rcabench.openapi.api.injections_api import InjectionsApi
from rcabench.openapi.api.traces_api import TracesApi
from rcabench.openapi.models.batch_evaluate_datapack_req import BatchEvaluateDatapackReq
from rcabench.openapi.models.container_ref import ContainerRef
from rcabench.openapi.models.datapack_result import DatapackResult
from rcabench.openapi.models.evaluate_datapack_spec import EvaluateDatapackSpec
from rcabench.openapi.models.event_type import EventType
from rcabench.openapi.models.stream_event import StreamEvent
from rcabench.openapi.models.submit_injection_req import SubmitInjectionReq
from sse_connection import listen_trace_events

from src.common.common import console


@pytest.fixture(scope="session")
def evaluations_api(setup_environment: ApiClient) -> EvaluationsApi:
    return EvaluationsApi(setup_environment)


@pytest.fixture(scope="session")
def injections_api(setup_environment: ApiClient) -> InjectionsApi:
    return InjectionsApi(setup_environment)


@pytest.fixture(scope="session")
def traces_api(setup_environment: ApiClient) -> TracesApi:
    """Fixture to provide TracesApi instance for testing."""
    return TracesApi(setup_environment)


class TestWorkflow:
    # Class variable to share datapack between tests
    shared_datapack: ClassVar[str | None] = None
    shared_execution_id: ClassVar[str | None] = None

    @pytest.mark.k8s
    @pytest.mark.order(1)
    def test_full_pipeline(
        self,
        evaluations_api: EvaluationsApi,
        injections_api: InjectionsApi,
        traces_api: TracesApi,
        case: dict[str, Any],
    ):
        """
        Test full pipeline:
            restart pedestal -> submit injection -> build datapack -> execute algorithm -> collect results

        Args:
            injections_api: InjectionsApi fixture
            traces_api: TracesApi fixture
            case: Test case data from test_injections_api.json, structure:
                {
                    "name": "test_case_name",
                    "description": "...",
                    "input": { ... },  # Request payload
                    "expected": {      # Optional expected values
                        "min_events": 1,
                        "timeout": 300
                    }
                }
        """
        console.print(f"\nRunning test case: {case.get('name', 'unnamed')}]")
        console.print(f"Description: {case.get('description', 'N/A')}")

        # Validate input data
        request_data = case.get("input")
        if request_data is None:
            pytest.fail("Test case must have 'input' field with request data")
        if not isinstance(request_data, dict):
            pytest.fail("Input data must be a dictionary")

        # Parse expected values
        expected = case.get("expected", {})
        timeout_seconds = expected.get("timeout", 300)  # Default 5 minutes
        min_events = expected.get("min_events", 1)

        # Create request body
        request_body = SubmitInjectionReq.from_dict(request_data)
        if request_body is None:
            pytest.fail("Failed to create SubmitInjectionReq from input data")

        # Submit the fault injection
        console.print("[bold blue]Submitting fault injection...[/bold blue]")
        response = injections_api.inject_fault(body=request_body)

        assert response is not None, "Response should not be None"
        assert response.code == 200, f"Expected status code 200, got {response.code}"
        assert response.data is not None, "Response data should not be None"
        assert response.data.items is not None, "Response items should not be None"
        assert len(response.data.items) > 0, "Response should containst one item"

        item = response.data.items[0]
        assert item.trace_id is not None, "Trace ID should not be None"
        assert isinstance(item.trace_id, str), "Trace ID should be a string"
        assert len(item.trace_id) > 0, "Trace ID should not be empty"

        console.print(
            "[bold green]✅ Fault injection submitted successfully[/bold green]"
        )
        console.print(f"\n[bold green]Trace ID: {item.trace_id}[/bold green]")

        # Listen to trace events and validate
        results = listen_trace_events(
            traces_api,
            trace_id=item.trace_id,
            timeout_seconds=timeout_seconds,
            min_events=min_events,
        )

        console.print()
        post_process_events(evaluations_api, case, results)


# class TestTraces:
#     def test_listen_trace_events(
#         self,
#         evaluations_api: EvaluationsApi,
#         traces_api: TracesApi,
#         case: dict[str, Any],
#     ):
#         # Parse expected values
#         expected = case.get("expected", {})
#         timeout_seconds = expected.get("timeout", 300)  # Default 5 minutes
#         min_events = expected.get("min_events", 1)

#         results = listen_trace_events(
#             traces_api,
#             trace_id="6d8b7e08-32d7-4e55-8a80-303daf8f9288",
#             timeout_seconds=timeout_seconds,
#             min_events=min_events,
#         )

#         post_process_events(evaluations_api, case, results)


def post_process_events(
    evaluations_api: EvaluationsApi, case: dict[str, Any], results: list[str]
) -> None:
    """
    Post-process trace events to extract datapack and evaluate algorithms.

    Args:
        evaluations_api: API client for algorithm evaluation
        case: Test case containing algorithm configurations
        results: Raw SSE event data from trace stream
    """
    # Extract algorithm configurations from test case
    algorithm_configs: list[dict[str, Any]] = case["input"].get("algorithms", [])
    if not algorithm_configs:
        console.print(
            "[bold yellow]⚠️ No algorithms specified, skipping evaluation[/bold yellow]"
        )
        return

    # Parse raw SSE data into StreamEvent objects
    parsed_events: list[StreamEvent] = []
    for raw_event_data in results[:-1]:
        stream_event = StreamEvent.from_json(raw_event_data)
        if stream_event is not None:
            parsed_events.append(stream_event)

    # Group events by event type for efficient lookup
    events_by_type: dict[EventType, list[StreamEvent]] = {}
    for stream_event in parsed_events:
        assert stream_event.event_name is not None, "Event name should not be None"
        events_by_type.setdefault(stream_event.event_name, []).append(stream_event)

    # Verify algorithm execution completed
    if EventType.AlgoResultCollection not in events_by_type:
        console.print(
            "[bold yellow]⚠️ No AlgoResultCollection event found in the trace events[/bold yellow]"
        )
        return

    # Extract datapack identifier from build success event
    datapack_identifier: str = ""
    if EventType.DatapackBuildSucceed in events_by_type:
        build_success_events = events_by_type[EventType.DatapackBuildSucceed]
        assert len(build_success_events) == 1, (
            f"Expected exactly one DatapackBuildSucceed event, found {len(build_success_events)}"
        )

        build_event = build_success_events[0]
        assert build_event.payload is not None, (
            "DatapackBuildSucceed event payload is None"
        )

        build_result = DatapackResult.from_dict(build_event.payload)
        assert build_result is not None, (
            "Failed to parse DatapackResult from event payload"
        )

        assert build_result.datapack is not None, (
            "Datapack identifier is None, cannot proceed with evaluation"
        )
        datapack_identifier = build_result.datapack
        console.print(
            f"\n[bold green]Extracted datapack: {datapack_identifier}[/bold green]"
        )

    # Parse algorithm container references
    algorithm_references: list[ContainerRef] = []
    for config in algorithm_configs:
        container_ref = ContainerRef.from_dict(config)
        if container_ref is not None:
            algorithm_references.append(container_ref)

    console.print(
        f"[bold blue]Preparing evaluation for {len(algorithm_references)} algorithm(s)[/bold blue]"
    )

    # Submit batch evaluation request
    console.print("[bold blue]Submitting algorithm evaluation request...[/bold blue]")

    evaluation_response = evaluations_api.evaluate_algorithm_on_datapacks(
        request=BatchEvaluateDatapackReq(
            specs=[
                EvaluateDatapackSpec(
                    algorithm=algorithm_ref,
                    datapack=datapack_identifier,
                )
                for algorithm_ref in algorithm_references
            ]
        )
    )

    # Validate evaluation response
    assert evaluation_response is not None, "Evaluation response is None"
    assert evaluation_response.code == 200, (
        f"Expected status code 200, got {evaluation_response.code}"
    )
    assert evaluation_response.data is not None, "Evaluation response data is None"

    console.print(
        "[bold green]✅ Algorithm evaluation completed successfully[/bold green]"
    )
    console.print("\n[bold blue]Evaluation Results:[/bold blue]")
    console.print(evaluation_response.data.to_dict())
