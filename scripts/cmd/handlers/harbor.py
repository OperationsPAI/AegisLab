from typing import Dict, List
from docker.errors import APIError, BuildError
from rich.console import Console
from rich.progress import track
from rich.prompt import Prompt
from . import ALGORITHMS_DIR, BENCHMARKS_DIR
from .console import select_choice, select_function
import docker
import os

TASK_CONFIGS = {
    "algorithm": {"dir": ALGORITHMS_DIR, "Dockerfile": "builder.Dockerfile"},
    "dataset": {"dir": BENCHMARKS_DIR, "Dockerfile": "Dockerfile"},
}


class Executor:
    def __init__(self, console: Console, config: Dict[str, str]):
        self.console = console
        self.config = config

        self.client = docker.from_env()

    def connect_harbor(self) -> None:
        """登录到 Harbor 仓库"""
        host = self.config["host"]
        try:
            self.client.login(
                username=self.config["username"],
                password=self.config["password"],
                registry=host,
            )
            self.console.print(
                f"Logged in to {host} [bold green]successfully[/bold green]."
            )
        except APIError as e:
            self.console.print(f"[bold red]Failed[/bold red] to login to {host}: {e}")

    def _build_images(
        self, task_type: str, image_configs: List[Dict[str, str]]
    ) -> None:
        """构建镜像"""
        self.images = []

        for image_config in track(image_configs, description="Building image..."):
            name = image_config["name"]
            tag = image_config.get("tag", "latest")
            repository = self.config.get("repository")
            repository = f"{repository}/{name}"
            task_config = TASK_CONFIGS[task_type]

            try:
                self.client.images.build(
                    path=".",
                    tag=f"{repository}:{tag}",
                    dockerfile=os.path.join(
                        task_config["dir"], name, task_config["Dockerfile"]
                    ),
                )
                self.console.print(
                    f"Image {repository}:{tag} built [bold green]successfully[/bold green]."
                )

                self.images.append({"repository": repository, "tag": tag})
            except BuildError as e:
                self.console.print(
                    f"[bold red]Failed[/bold red] to build image {repository}:{tag}: {e}"
                )
                break

    def push_images(self, task_type: str, image_configs: List[Dict[str, str]]) -> None:
        self._build_images(task_type, image_configs)

        for image in track(self.images, description="Pushing image to Harbor..."):
            repository = image["repository"]
            tag = image["tag"]

            try:
                push_logs = self.client.images.push(
                    repository,
                    tag=tag,
                    stream=True,
                    decode=True,
                )

                for log in push_logs:
                    if "aux" in log:
                        self.console.print(f"辅助信息: {log['aux']}")

                self.console.print(
                    f"Image {repository}:{tag} pushed to harbor [bold green]successfully[/bold green]."
                )
            except docker.APIError as e:
                self.console.print(
                    f"[bold red]Failed[/bold red] to push image {repository}:{tag}: {e}"
                )
                continue


def run_harbor(console: Console, args, harbor_config: Dict):
    @select_function
    def main(config: Dict):
        func = config["func"]
        kwargs = {}
        params = config.get("params", None)

        if func == executor.push_images:
            task_type = select_choice(console, "任务类型", list(TASK_CONFIGS.keys()))
            kwargs = {"task_type": task_type, "image_configs": []}

            if params:
                item = {}
                flag = True

                while flag:
                    for param in params:
                        name = param["name"]
                        item[name] = Prompt.ask(
                            f"请输入 {name} 的值",
                            choices=param.get("choices", None),
                            default=param.get("default", None),
                            show_choices="choices" in list(param.keys()),
                        )

                    kwargs["image_configs"].append(item)
                    flag = select_choice(console, "是否继续", ["Y", "N"]) == "Y"

        func(**kwargs)

    executor = Executor(console, harbor_config)
    executor.connect_harbor()

    if args.default2:
        task_type = "algorithm"
        image_configs = [{"name": "e-diagnose"}]
        executor.push_images(task_type, image_configs)

    runner_config = {
        "PushImages": {
            "func": executor.push_images,
            "params": [
                {
                    "name": "name",
                    "help": "镜像名称",
                    "default": "e-diagnose",
                },
                {
                    "name": "tag",
                    "help": "镜像tag",
                    "default": "latest",
                },
            ],
        },
    }

    main(runner_config)
