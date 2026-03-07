import json
import re
import shutil
import sys
from pathlib import Path

from python_on_whales import docker

from src.common.common import PROJECT_ROOT, ScopeType, console, settings
from src.formatter import PythonFormatter
from src.swagger.common import SWAGGER_ROOT, Generator


class PythonSDK(Generator):
    """Class to generate Python SDK from Swagger JSON using OpenAPI Generator."""

    PYTHON_SDK_DIR = PROJECT_ROOT / "sdk" / "python"
    PYTHON_SDK_GEN_DIR = PROJECT_ROOT / "sdk" / "python-gen"
    PYTHON_GENERATOR_CONFIG_DIR = PROJECT_ROOT / ".openapi-generator" / "python"

    def __init__(self, version: str) -> None:
        self.version = version

    def _update_version(self) -> None:
        """
        Update version information in various project files.
            - sdk/python/src/rcabench/__init__.py
            - src/config.dev.toml
            - helm/values.yaml
        """
        # Update version in sdk/python/src/rcabench/__init__.py
        init_python_path = self.PYTHON_SDK_DIR / "src" / "rcabench" / "__init__.py"
        with open(init_python_path, encoding="utf-8") as f:
            init_content = f.read()

        pattern = r'__version__\s*=\s*["\'].*?["\']'
        replacement = f'__version__ = "{self.version}"'
        new_init_content = re.sub(pattern, replacement, init_content)

        with open(init_python_path, "w", encoding="utf-8") as f:
            f.write(new_init_content)

    def generate(self) -> None:
        """
        Generate Python SDK from Swagger JSON using OpenAPI Generator.

        Post-process the generated SDK to adjust package structure and formatting.
        """
        # 1. Update generator config with the specified version
        sdk_config = self.PYTHON_GENERATOR_CONFIG_DIR / "config.json"
        with open(sdk_config) as f:
            config_data = json.load(f)

        config_data["packageVersion"] = self.version

        tmp_sdk_config = self.PYTHON_GENERATOR_CONFIG_DIR / "config_tmp.json"
        with open(tmp_sdk_config, "w") as f:
            json.dump(config_data, f, indent=2)

        console.print(
            f"[bold green]✅ Updated packageVersion to {self.version}[/bold green]"
        )

        # 2. Generate SDK using OpenAPI Generator
        console.print("[bold blue]Step 1: Generating Python SDK...[/bold blue]")

        if self.PYTHON_SDK_GEN_DIR.exists():
            shutil.rmtree(self.PYTHON_SDK_GEN_DIR)

        self.PYTHON_SDK_GEN_DIR.mkdir(parents=True)

        volume_path = Path("/local")
        relative_swagger = SWAGGER_ROOT.relative_to(PROJECT_ROOT)
        relative_sdk_gen = self.PYTHON_SDK_GEN_DIR.relative_to(PROJECT_ROOT)
        relative_generator_config = self.PYTHON_GENERATOR_CONFIG_DIR.relative_to(
            PROJECT_ROOT
        )

        container_input_path = volume_path / relative_swagger / "converted" / "sdk.json"
        container_output_path = volume_path / relative_sdk_gen
        container_config_path = (
            volume_path / relative_generator_config / "config_tmp.json"
        )
        container_templates_path = volume_path / relative_generator_config / "templates"

        # Get current user UID and GID to avoid permission issues
        import os

        current_user = os.getuid()
        current_group = os.getgid()

        try:
            docker.run(
                settings.generator_image,
                command=[
                    "generate",
                    "-i",
                    container_input_path.as_posix(),
                    "-g",
                    "python",
                    "-o",
                    container_output_path.as_posix(),
                    "-c",
                    container_config_path.as_posix(),
                    "-t",
                    container_templates_path.as_posix(),
                    "--git-host",
                    "github.com",
                    "--git-repo-id",
                    "AegisLab",
                    "--git-user-id",
                    "OperationsPAI",
                ],
                volumes=[(PROJECT_ROOT, volume_path)],
                user=f"{current_user}:{current_group}",
                remove=True,
            )
        except Exception as e:
            console.print(
                f"[bold_red]❌ Error during python sdk generation: {e}[/bold_red]"
            )
            sys.exit(1)
        finally:
            if tmp_sdk_config.exists():
                tmp_sdk_config.unlink(missing_ok=True)

        console.print(
            "[bold green]✅ Original python SDK generated successfully![/bold green]"
        )
        console.print()

        # 3. Post-process generated SDK (if any post-processing is needed)
        console.print("[bold blue]Step 2: Post-processing generated SDK...[/bold blue]")

        dst = self.PYTHON_SDK_DIR / "src" / "rcabench" / "openapi"
        python_sdk_pyproject = self.PYTHON_SDK_DIR / "pyproject.toml"

        if dst.exists():
            shutil.rmtree(dst)
        if python_sdk_pyproject.exists():
            python_sdk_pyproject.unlink(missing_ok=True)

        shutil.copytree(self.PYTHON_SDK_GEN_DIR / "openapi", dst)
        shutil.copyfile(
            self.PYTHON_SDK_GEN_DIR / "pyproject.toml", python_sdk_pyproject
        )

        old_str = "openapi"
        new_str = "rcabench.openapi"

        for filepath in dst.rglob("*.py"):
            if filepath.is_file():
                content = filepath.read_text(encoding="utf-8")
                new_content = content.replace(old_str, new_str)
                if new_content != content:
                    filepath.write_text(new_content, encoding="utf-8")

        py_typed_files = dst / "py.typed"
        if py_typed_files.exists():
            py_typed_files.unlink(missing_ok=True)

        console.print(
            "[bold green]✅ Python SDK post procession completed successfully![/bold green]"
        )
        console.print()

        # 4. Format the generated SDK code
        console.print(
            "[bold blue]Step 3: Formatting post-processed Python SDK...[/bold blue]"
        )
        formatter = PythonFormatter(scope=ScopeType.SDK)
        formatter.run()

        # 5. Update version information in project files
        self._update_version()
