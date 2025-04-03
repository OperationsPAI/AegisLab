from typing import Dict
from datetime import datetime
from rcabench.logger import CustomLogger
from rcabench.rcabench import RCABenchSDK, Node
import asyncio
import json
import math
import os
import random
from copy import deepcopy


PARENT_DIR = os.path.dirname(os.path.abspath(__file__))
CONFIG_NAME = "config/{env}.json"

FINISH_EVENT = asyncio.Event()
LOCK = asyncio.Lock()
GROUP_ID = None

logger = CustomLogger().logger
random.seed(42)


async def periodic_task(config: Dict, func) -> None:
    """协程1：每隔指定时间执行脚本"""
    try:
        for i in range(config["n_trial"] + 1):
            interval = config["interval"]
            log_msg = "Executing script..." + (
                f" (next in {interval}min)" if i < config["n_trial"] else ""
            )
            logger.info(log_msg)

            await func(config["command"])
            await asyncio.sleep(interval * 60)

    finally:
        FINISH_EVENT.set()
        logger.info("Periodic Task Completed")


async def delayed_request(config: Dict):
    """协程2：延迟指定时间后发送请求，然后结束"""
    delay_time = config["interval"] - config["pre_duration"] - config["fault_duration"]
    await asyncio.sleep(delay_time * 60)

    logger.info(f"Send request (trigger after {delay_time} minutes)")
    try:
        data = await execute_injection(config)
        if data:
            async with LOCK:
                global GROUP_ID
                GROUP_ID = data.get("group_id")

    except Exception as e:
        logger.error(f"Request failed: {str(e)}")
        return "error"

    finally:
        logger.info("Delayed Request Task Completed")


async def run_deploy_command(command: str) -> None:
    process = await asyncio.create_subprocess_shell(
        command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )

    await asyncio.gather(
        read_stream(process.stdout, "STDOUT"),
        read_stream(process.stderr, "STDERR"),
    )

    return_code = await process.wait()
    if return_code == 0:
        logger.info("Deploy successfully")
    else:
        logger.error(f"Deploy failed, return_code: {return_code}")


async def read_stream(stream, prefix):
    while not stream.at_eof():
        line = await stream.readline()
        if line:
            print(f"[{prefix}] {line.decode().strip()}")


async def execute_injection(config: Dict) -> Dict[str, any]:
    sdk = RCABenchSDK(config["base_url"])

    injection_params = sdk.injection.get_parameters()
    if not injection_params:
        logger.error("Injection Params invalid")
        return None

    payloads = []
    for _ in range(config["n_trial"]):
        new_payload = deepcopy(injection_params)
        payloads.append(generate_injection_dict(new_payload))

    body = {
        "interval": config["interval"],
        "pre_duration": config["pre_duration"],
        "benchmark": config["benchmark"],
        "payloads": payloads,
    }
    data = sdk.injection.submit(**body)

    req_path = os.path.join(config["output_path"], "request.json")
    logger.info(f"Request params store in {req_path}")
    with open(req_path, "w") as f:
        json.dump(body, f, indent=4)

    resp_path = os.path.join(config["output_path"], "response.json")
    logger.info(f"Response store in {req_path}")
    with open(resp_path, "w") as f:
        json.dump(data, f, indent=4)

    return data


def generate_injection_dict(spec: Node) -> Node:
    def fill_node(node: Node):
        if "children" in node:
            for children, sub_node in node["children"].items():
                fill_node(sub_node)
        if "children" not in node:
            node["value"] = random.randint(node["range"][0], node["range"][1])

    chosen_key = random.choice(list(spec["children"].keys()))
    fill_node(spec["children"][chosen_key])
    spec["value"] = chosen_key
    return spec


def download_datasets(config: Dict[str, any]) -> None:
    sdk = RCABenchSDK(config["base_url"])
    sdk.dataset.download([GROUP_ID], config["output_path"])
    logger.info("Download datasets successfully")


async def main(config: Dict[str, any]) -> None:
    asyncio.create_task(periodic_task(config, run_deploy_command))
    asyncio.create_task(delayed_request(config))

    await FINISH_EVENT.wait()

    download_datasets(config)


if __name__ == "__main__":
    env_mode = os.getenv("ENV_MODE", "dev")
    default_config_path = os.path.join(PARENT_DIR, CONFIG_NAME.format(env=env_mode))
    with open(os.getenv("CONFIG_FILE", default_config_path)) as f:
        config = json.load(f)

    default_output = os.path.join(
        PARENT_DIR, "output", datetime.now().strftime("%Y-%m-%d-%H:%M:%S")
    )
    config["output_path"] = default_output
    if not os.path.exists(default_output):
        os.makedirs(default_output)

    dynamic_params = {
        "command": os.getenv("COMMAND"),
        "namespace": os.getenv("NAMESPACE"),
        "services": os.getenv("SERVICES", "").split(","),
    }
    config.update({k: v for k, v in dynamic_params.items() if v})

    FINISH_EVENT.clear()
    try:
        asyncio.run(main(config))
    except KeyboardInterrupt:
        logger.info("Program has been manually terminated")
