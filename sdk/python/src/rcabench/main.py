from typing import Any, Dict, List
from .logger import CustomLogger
from .rcabench import RCABenchSDK
import asyncio
import os
import random

CustomLogger()
logger = CustomLogger().logger


class InjectHelper:
    def __init__(self, specification, keymap):
        self.specification = specification
        self.keymap = keymap

    def generate_injection_dict(self):
        fault_type_name = random.choice(list(self.specification.keys()))
        # Find the fault_type key from the keymap
        fault_type_key = None
        for key, value in self.keymap.items():
            if value == fault_type_name:
                fault_type_key = key
                break

        if not fault_type_key:
            raise ValueError(f"Fault type {fault_type_name} not found in keymap.")

        # Get the specification for the selected fault type
        spec = self.specification.get(fault_type_name)
        if not spec:
            raise ValueError(f"Specification for {fault_type_name} not found.")

        # Construct the inject_spec by generating random values within the constraints
        inject_spec = {}
        for field in spec:
            field_name = field["FieldName"]
            min_val = field["Min"]
            max_val = field["Max"]
            inject_spec[field_name] = random.randint(min_val, max_val)

        # Return the constructed dictionary
        return {
            "fault_type": int(fault_type_key),
            "inject_spec": inject_spec,
        }


class Example:
    def __init__(self, url: str):
        self.server_url = url
        self.sdk = RCABenchSDK(self.server_url)

    def batch_execute_algorithm(self, payloads: List[Any]):
        data = self.sdk.algorithm.execute(payloads)
        if not data:
            return None

        return data

    def batch_inject(self, n_trial: int = 10):
        excluded_pods = ["mysql"]

        injection_params = self.sdk.injection.get_parameters()
        if not injection_params:
            return None

        logger.info(injection_params)
        helper = InjectHelper(
            specification=injection_params["specification"],
            keymap=injection_params["keymap"],
        )

        namespace_pod_info = self.sdk.injection.get_namespace_pod_info()
        if not namespace_pod_info:
            return None

        logger.info(namespace_pod_info)
        namespace = random.choice(list(namespace_pod_info["namespace_info"].keys()))

        payloads = []
        for _ in range(n_trial):
            pod = random.choice(namespace_pod_info.namespace_info[namespace])
            params = helper.generate_injection_dict()
            while params["fault_type"] in excluded_pods:
                params = helper.generate_injection_dict()

            payloads.append(
                {
                    "faultType": params["fault_type"],
                    "duration": random.randint(5, 10),
                    "injectNamespace": namespace,
                    "injectPod": pod,
                    "spec": params["inject_spec"],
                    "benchmark": "clickhouse",
                }
            )

        logger.info(payloads)

        data = self.sdk.injection.execute(payloads)
        if not data:
            return None

        return data

    def delete_dataset(self, ids: List[int]):
        self.sdk.dataset.delete(ids)

    def list_algorithm(self):
        data = self.sdk.algorithm.list()
        return data["algorithms"] if data else None

    def list_dataset(self, page_num: int, page_size: int):
        data = self.sdk.dataset.list(page_num, page_size)
        return data["datasets"] if data else None

    def list_injection(self):
        return self.sdk.injection.list()

    async def execute_algorithm_and_collection(self, payloads: List[Any]) -> Any:
        data = self.sdk.algorithm.execute(payloads)
        if not data:
            return None

        logger.info(data)

        report = await self.sdk.start_multiple_stream(
            data["task_ids"],
            url="/algorithms/{task_id}/stream",
            keyword="execution_id",
            timeout=30,
        )
        logger.info(report)
        return report

    async def execute_injection_and_building(
        self, payloads: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        data = self.sdk.injection.execute(payloads)
        if not data:
            return None

        logger.info(data)

        report = await self.sdk.start_multiple_stream(
            data["task_ids"], url="/injections/{task_id}/stream", keyword="dataset"
        )
        logger.info(report)
        return data["group_id"], report

    async def test_workflow(
        self, injection_payloads: List[Any], algorithms: List[Any]
    ) -> None:
        data = await self.execute_injection_and_building(injection_payloads)
        if not data:
            return None

        group_id, report = data
        datasets = []
        for _, message in report["results"].items():
            datasets.append(message["dataset"])

        self.sdk.task_manager.cleanup()

        detector_payloads = []
        for dataset in datasets:
            detector_payloads.append(
                {
                    "benchmark": "clickhouse",
                    "algorithm": "detector",
                    "dataset": dataset,
                },
            )

        data = self.batch_execute_algorithm(detector_payloads)
        if not data:
            return data
        logger.info(data)

        algorithm_payloads = []
        for dataset in datasets:
            for algorithm in algorithms:
                algorithm_payloads.append(
                    {
                        "benchmark": "clickhouse",
                        "algorithm": algorithm,
                        "dataset": dataset,
                    },
                )

        report = await self.execute_algorithm_and_collection(algorithm_payloads)
        execution_ids = []
        for _, message in report["results"].items():
            execution_ids.append(message["execution_id"])

        self.sdk.dataset.download([group_id], os.getcwd())

        evaluation_params = {"execution_ids": execution_ids, "algorithms": algorithms}
        data = self.sdk.evaluation.execute(evaluation_params)
        if not data:
            return None

        logger.info(data["results"])

        return datasets


if __name__ == "__main__":
    url = "http://localhost:8082"
    example = Example(url)
    logger.info(example.list_algorithm())
    logger.info(example.list_dataset(3, 20))

    injection_payloads = [
        {
            "duration": 1,
            "faultType": 5,
            "injectNamespace": "ts",
            "injectPod": "ts-preserve-service",
            "spec": {"CPULoad": 1, "CPUWorker": 3},
            "benchmark": "clickhouse",
        }
    ]
    algorithms = ["e-diagnose"]
    try:
        asyncio.run(example.test_workflow(injection_payloads, algorithms))
    except KeyboardInterrupt:
        print("Shutting down...")
