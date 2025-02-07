import unittest
from rcabench import RCABenchSDK, RunAlgorithmPayload
from pprint import pprint
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


class TestRCABenchSDK(unittest.TestCase):
    def __init__(self, methodName="runTest"):
        super().__init__(methodName)
        self.base_url = "http://10.10.10.220:32080"
        self.sdk = RCABenchSDK(self.base_url)

    def test_get_algorithms(self):
        print(self.sdk.get_algorithms())

    def test_submit_injection(self):
        base_url = "http://10.10.10.220:32080"  # 替换为实际服务器地址
        sdk = RCABenchSDK(base_url)

        injection_params = sdk.get_injection_parameters()
        helper = InjectHelper(
            specification=injection_params.specification, keymap=injection_params.keymap
        )
        params = helper.generate_injection_dict()

        namespace_pod_info = sdk.get_injection_namespace_pod_info()
        namespace = random.choice(list(namespace_pod_info.namespace_info.keys()))
        pod = random.choice(namespace_pod_info.namespace_info[namespace])

        faults = [
            {
                "faultType": params["fault_type"],
                "duration": random.randint(5, 10),
                "injectNamespace": namespace,
                "injectPod": pod,
                "spec": params["inject_spec"],
                "benchmark": "clickhouse",
            }
        ]
        task_response = sdk.inject(faults)
        pprint(faults)
        pprint(task_response)

        # for task_id in task_response["data"]:
        #     status_response = sdk.get_task_status(task_id)
        #     pprint(status_response)

        #     details = sdk.get_task_details(task_id)
        #     pprint(details)

    def test_submit_evaluation(self):
        base_url = "http://localhost:8082"
        sdk = RCABenchSDK(base_url)
        algo_bench = sdk.get_algo_bench()
        pprint(algo_bench)
        datasets = sdk.get_datasets()
        pprint(datasets)
        task_response = sdk.submit_task(
            "RunAlgorithm",
            RunAlgorithmPayload(
                algorithm="metis",  # random.choice(algo_bench.algorithms)
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


if __name__ == "__main__":
    TestRCABenchSDK().test_submit_injection()
