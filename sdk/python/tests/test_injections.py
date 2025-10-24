# Run this file:
# uv run pytest -s tests/test_injections.py
from pprint import pprint

import pytest

from rcabench.openapi.api import InjectionApi
from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.models import (
    DtoAlgorithmItem,
    DtoLabelItem,
    DtoSubmitInjectionReq,
    DtoSubmitInjectionResp,
    HandlerNode,
)


class TestInjections:
    @pytest.fixture(autouse=True)
    def _setup(self, rcabench_client: ApiClient):
        self.injection_api = InjectionApi(rcabench_client)

    @pytest.mark.parametrize(
        "container_idx, cpu_load, cpu_worker",
        [
            (6, 1, 3),
        ],
    )
    @pytest.mark.k8s
    @pytest.mark.v1
    def test_inject_one_fault(self, container_idx: int, cpu_load: int, cpu_worker: int):
        specs = [
            HandlerNode(
                children={
                    "4": HandlerNode(
                        children={
                            "0": HandlerNode(value=1),
                            "1": HandlerNode(value=0),
                            "2": HandlerNode(value=container_idx),
                            "3": HandlerNode(value=cpu_load),
                            "4": HandlerNode(value=cpu_worker),
                        },
                    )
                },
                value=4,
            )
        ]

        resp = self.injection_api.api_v1_injections_post(
            body=DtoSubmitInjectionReq(
                algorithms=[
                    DtoAlgorithmItem(name="traceback"),
                ],  # Algorithm execution container name
                benchmark="clickhouse",  # Data collection container name
                container_name="ts_cn",  # Pedestal container name
                container_tag="v1.0.0-213-gf9294111",  # Pedestal container tag
                interval=2,  # The whole period in minutes
                project_name="pair_diagnosis",
                pre_duration=1,  # Normal time in minutes
                specs=specs,
                labels=[
                    DtoLabelItem(key="env", value="testing"),
                    DtoLabelItem(key="batch", value="bootstrap"),
                ],
            )
        )

        assert resp is not None
        assert resp.data is not None and isinstance(resp.data, DtoSubmitInjectionResp)

        pprint(resp.data)
