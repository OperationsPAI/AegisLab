import time

import pytest
from rcabench.openapi.api.traces_api import TracesApi

from src.common.common import console


def listen_trace_events(
    traces_api: TracesApi, trace_id: str, timeout_seconds: int, min_events: int
) -> list[str]:
    # Listen to trace events via SSE with proper error handling
    results: list[str] = []
    sse_client = None
    try:
        sse_client = traces_api.get_trace_events(trace_id=trace_id)
        assert sse_client is not None, "SSE client should not be None"

        console.print(
            f"[bold blue]Listening to trace events (timeout: {timeout_seconds}s)...[/bold blue]"
        )

        start_time = time.time()
        event_count = 0
        received_end = False

        for event in sse_client.events():
            # Check timeout
            elapsed = time.time() - start_time
            if elapsed > timeout_seconds:
                pytest.fail(
                    f"Timeout after {timeout_seconds}s waiting for trace events. "
                    f"Received {event_count} events."
                )

            event_count += 1
            event_data = event.data
            event_type = event.event
            results.append(event_data)
            console.print(f"  [{event_count}] {event_type}: {event_data}")

            # Handle different event types
            if event_type == "end":
                console.print("\n[bold green]✅ Received 'end' event[/bold green]")
                received_end = True
                break

        # Verify we received events
        assert event_count >= min_events, (
            f"Expected at least {min_events} events, but received {event_count}"
        )

        # Verify we received the end event
        assert received_end, (
            f"Did not receive 'end' event after {event_count} events. "
            f"This may indicate the trace is still running or an error occurred."
        )

        console.print(
            f"[bold green]✅ Listened to trace events completed successfully "
            f"({event_count} events in {time.time() - start_time:.2f}s)[/bold green]"
        )

        return results

    finally:
        # Ensure SSE connection is always closed
        if sse_client is not None:
            try:
                sse_client.close()
                console.print("[dim]SSE connection closed[/dim]")
            except Exception as e:
                console.print(
                    f"[yellow]Warning: Error closing SSE client: {e}[/yellow]"
                )
