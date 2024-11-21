import toml
from pathlib import Path
import time
import requests
import json
import pandas as pd
import concurrent.futures

# your key
api_key = ""

def get_one_input(file_path):
    with open(file_path, 'r') as file:
        data = toml.load(file)

    top_events = {
        "log_events": data.get("log_events", [])[:5],
        "trace_events": data.get("trace_events", [])[:5],
        "metric_events": data.get("metric_events", [])[:5],
    }

    input_data = "Input: " + str(top_events) + "\n" + "Output:"
    return input_data

def get_instruction_v1():
    prompt = """
    Instruction: You are an expert in root cause analysis.

    Task: Based on the provided events (log_events, trace_events, and metric_events), analyze which service experienced a failure and identify the type of fault.

    Fault types to choose from:
    - cpu-exhaustion
    - memory-exhaustion
    - pod-failure

    Output format:
    Provide the analysis in a structured JSON format like this:
    {
      "Analysis": "<Detailed explanation of the findings>",
      "Root cause service": "<Service name>",
      "Fault type": "<Fault type>"
    }

    Example Input: {'log_events': [...], 'trace_events': [...], 'metric_events': [...]}

    Example Output:
    {
      "Analysis": "The ts-route-service is experiencing a cpu-exhaustion issue, given the extreme increase in CPU usage and CPU limit utilization. The ts-inside-payment-service pod and ts-train-service pod appear to have performance issues related to slow durations but don't seem to indicate a direct resource exhaustion type failure like CPU or memory exhaustion.",
      "Root cause service": "ts-route-service",
      "Fault type": "cpu-exhaustion"
    }
    """
    return prompt.strip()

def get_instruction():
    prompt = """
    Instruction: You are an expert in root cause analysis.

    Task: Based on the provided events (log_events, trace_events, and metric_events), analyze which service experienced a failure and identify the type of fault. Make sure to consider both resource utilization patterns and service-level behaviors revealed through trace and log events.

    Fault types to choose from:
    - cpu-exhaustion: Indicated by sustained high CPU usage in metric events.
    - memory-exhaustion: Indicated by high memory usage or memory limit violations in metric events.
    - pod-failure: Identified through trace events showing significant delays or errors in critical paths, especially for a specific pod or service, along with potential confirmation from metric events.

    Output format:
    Provide the analysis in a structured JSON format like this:

    {
    "Analysis": "<Detailed explanation of the findings>",
    "Root cause service": "<Service name or pod>",
    "Fault type": "<Fault type>"
    }

    Example Input:
    {'log_events': [...], 'trace_events': [...], 'metric_events': [...]}

    Example Output 1:
    {
    "Analysis": "The ts-travel2-service pod appears to have experienced a pod-failure. Trace events show a significant latency increase in BasicErrorController.error (24943.26% over normal) and cyclic dependencies in its parent service. Metric events confirm abnormal resource usage, but the root cause is likely the pod's failure to handle internal requests.",
    "Root cause service": "ts-travel2-service",
    "Fault type": "pod-failure"
    }

    Example Output 2:
    {
      "Analysis": "The ts-route-service is experiencing a cpu-exhaustion issue, given the extreme increase in CPU usage and CPU limit utilization. The ts-inside-payment-service pod and ts-train-service pod appear to have performance issues related to slow durations but don't seem to indicate a direct resource exhaustion type failure like CPU or memory exhaustion.",
      "Root cause service": "ts-route-service",
      "Fault type": "cpu-exhaustion"
    }
    """
    return prompt.strip()


def chat_with_gpt(instruction,input_data):
    url = "https://aigptx.top/v1/chat/completions"
    headers = {
        # "Content-Type": "application/json",
        "Authorization": 'Bearer ' + api_key
    }

    data = {
        # change model here
        # "model": "gpt-4",
        "model": "gpt-3.5-turbo", 
        "messages": [
            {
                "role": "system",
                "content": instruction
            },
            {
                "role": "user",
                "content": input_data
            }
        ]
    }
    # print("data: \n",data)
    response = requests.post(url, json=data, headers=headers, stream=False)
    res = response.json()
    output = res['choices'][0]['message']['content']

    return output


def process_events_files(root_base_dir):
    outputs = {}

    # 遍历 root_base_dir 下所有子文件夹
    for folder in root_base_dir.iterdir():
        if folder.is_dir():
            events_file_path = folder / "events.toml"
            if events_file_path.exists():
                try:
                    # 处理 events.toml 文件
                    instruction = get_instruction()
                    input_data = get_one_input(events_file_path)
                    output = chat_with_gpt(instruction, input_data)

                    # 将结果存储到 outputs 中
                    output_dict = json.loads(output)
                    outputs[folder.name] = output_dict
                except Exception as e:
                    print(f"Error processing {events_file_path}: {e}")

    return outputs

def read_ground_truth(fault_injection_file):
    """
    读取 fault_injection.toml 并解析 ground truth 数据。
    """
    data = toml.load(fault_injection_file)
    ground_truth = {}
    
    for injection in data.get("chaos_injection", []):
        case = injection["case"]
        service = injection["service"]
        chaos_type = injection["chaos_type"]
        ground_truth[case] = {"service": service, "fault_type": chaos_type}
    
    return ground_truth

def evaluate_accuracy(outputs, ground_truth):
    """
    分别计算 service (top1) 和 fault type 的准确率。
    """
    total_cases = len(outputs)
    if total_cases == 0:
        return {"service_accuracy": 0.0, "fault_type_accuracy": 0.0}

    correct_service = 0
    correct_fault_type = 0

    for case, output in outputs.items():
        ground_truth_case = ground_truth.get(case, {})
        if not ground_truth_case:
            continue

        predicted_service = output.get("Root cause service")
        predicted_fault_type = output.get("Fault type")
        
        if predicted_service == ground_truth_case["service"]:
            correct_service += 1
        
        if predicted_fault_type == ground_truth_case["fault_type"]:
            correct_fault_type += 1

    service_accuracy = correct_service / total_cases
    fault_type_accuracy = correct_fault_type / total_cases

    return {"service_accuracy": service_accuracy, "fault_type_accuracy": fault_type_accuracy}

def process_events_files_gpt4(root_base_dir):
    outputs = {}

    # 遍历 root_base_dir 下所有子文件夹
    for folder in root_base_dir.iterdir():
        if folder.is_dir():
            events_file_path = folder / "events.toml"
            if events_file_path.exists():
                try:
                    # 处理 events.toml 文件
                    instruction = get_instruction()
                    input_data = get_one_input(events_file_path)
                    raw_output = chat_with_gpt(instruction, input_data)

                    # 提取内容并解析 JSON
                    start_index = raw_output.find("{")
                    end_index = raw_output.rfind("}")
                    if start_index != -1 and end_index != -1:
                        output = json.loads(raw_output[start_index:end_index + 1])
                        outputs[folder.name] = output
                    else:
                        raise ValueError("Failed to parse GPT output as JSON")
                except Exception as e:
                    print(f"Error processing {events_file_path}: {e}")

    return outputs


# # run all cases
# if __name__ == "__main__":
#     root_base_dir = Path(r"E:\Project\Git\RCA_Dataset\test\ts")
#     outputs = process_events_files_gpt4(root_base_dir)
#     output_file = root_base_dir / "gpt_4_results.json"
#     with open(output_file, "w") as f:
#         json.dump(outputs, f, indent=4)

#     # for folder_name, output in outputs.items():
#     #     print(f"Folder: {folder_name}")
#     #     print(f"Output: {output}\n")
    
#     fault_injection_file = root_base_dir / "fault_injection.toml"
#     ground_truth = read_ground_truth(fault_injection_file)

#     # 计算准确率
#     accuracy_results = evaluate_accuracy(outputs, ground_truth)
#     print("Service Accuracy:", accuracy_results["service_accuracy"])
#     print("Fault Type Accuracy:", accuracy_results["fault_type_accuracy"])


# run one case
if __name__ == "__main__":
    root_base_dir = Path(r"E:\Project\Git\RCA_Dataset\test\ts\ts-consign-service-1027-1326")
    events_file_path = root_base_dir / "events.toml"

    instruction = get_instruction()
    input_data = get_one_input(events_file_path)
    output = chat_with_gpt(instruction,input_data)
    print(output)

