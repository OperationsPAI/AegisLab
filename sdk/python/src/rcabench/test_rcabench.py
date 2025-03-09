from .entity import RunAlgorithmPayload
from .rcabench import RCABenchSDK
from typing import List
from pprint import pprint
import asyncio
import random


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


class TestRCABenchSDK:
    def __init__(self, url: str):
        # 替换为实际服务器地址
        self.base_url = url
        self.sdk = RCABenchSDK(self.base_url)

    def test_get_algorithms(self):
        pprint(self.sdk.get_algorithms())

    def test_get_evaluations(self):
        base_url = "http://localhost:8082"
        sdk = RCABenchSDK(base_url)
        algo_bench = sdk.get_algo_bench()
        pprint(algo_bench)
        datasets = sdk.get_datasets()
        pprint(datasets)
        task_response = sdk.submit_task(
            "RunAlgorithm",
            RunAlgorithmPayload(
                algorithm="metis",
                benchmark=random.choice(algo_bench.benchmarks),
                dataset=random.choice(datasets.datasets),
            ),
        )
        pprint(task_response)

        task_id = task_response.taskID
        status_response = sdk.get_task_status(task_id)
        pprint(status_response)

        details = sdk.get_task_details(task_id)
        pprint(details)

    def test_submit_injection(self, n_trial: int = 10):
        excluded_pods = ["mysql"]

        injection_params = self.sdk.get_injection_parameters()
        helper = InjectHelper(
            specification=injection_params.specification, keymap=injection_params.keymap
        )

        namespace_pod_info = self.sdk.get_injection_namespace_pod_info()
        namespace = random.choice(list(namespace_pod_info.namespace_info.keys()))
        pprint(namespace_pod_info)

        faults = []
        for _ in range(n_trial):
            pod = random.choice(namespace_pod_info.namespace_info[namespace])
            params = helper.generate_injection_dict()
            while params["fault_type"] in excluded_pods:
                params = helper.generate_injection_dict()

            faults.append(
                {
                    "faultType": params["fault_type"],
                    "duration": random.randint(5, 10),
                    "injectNamespace": namespace,
                    "injectPod": pod,
                    "spec": params["inject_spec"],
                    "benchmark": "clickhouse",
                }
            )

        pprint(faults)

        task_response = self.sdk.submit_injection(faults)
        pprint(task_response)

        return task_response

    def test_injection_and_building(self) -> List[str]:
        injection_payload = [
            {
                "duration": 1,
                "faultType": 5,
                "injectNamespace": "ts",
                "injectPod": "ts-preserve-service",
                "spec": {"CPULoad": 1, "CPUWorker": 3},
                "benchmark": "clickhouse",
            }
        ]

        injection_resp = self.sdk.injection.submit_injection(injection_payload)
        pprint(injection_resp)

        datasets = self._get_results(
            injection_resp, self.sdk.get_injection_stream, "dataset"
        )
        pprint(datasets)

        return datasets

    async def test_algorithm_and_collection(
        self, algorithms: List[str], datasets: List[str]
    ) -> List[str]:
        algorithm_payload = []
        for dataset in datasets:
            for algorithm in algorithms:
                algorithm_payload.extend(
                    [
                        {
                            "benchmark": "clickhouse",
                            "algorithm": algorithm,
                            "dataset": dataset,
                        },
                    ]
                )

        algorithm_resp = self.sdk.algorithm.execute(algorithm_payload)
        pprint(algorithm_resp)

        task_ids = algorithm_resp["task_ids"]

        report = await self.sdk.start_multiple_stream(
            task_ids, url="/algorithms/{task_id}/stream", timeout=30
        )
        print(report)

    def test_pipeline(self, algorithms: List[str]) -> List[str]:
        self.sdk.evaluation.execute()


if __name__ == "__main__":
    url = "http://localhost:8082"
    sdk = TestRCABenchSDK(url)
    try:
        asyncio.run(
            sdk.test_algorithm_and_collection(
                algorithms=["e-diagnose", "e-diagnose"],
                datasets=["ts-ts-preserve-service-cpu-exhaustion-7wqhd5"],
            )
        )
    except KeyboardInterrupt:
        print("Shutting down...")
