import json
import os
import shutil
import sys
from pathlib import Path

from python_on_whales import docker

from src.common.command import run_command
from src.common.common import PROJECT_ROOT, console, settings
from src.swagger.common import SWAGGER_ROOT, RunMode


class TypeScriptSDK:
    """TypeScript generator for separate portal/admin audience specs."""

    SDK_ROOT_DIR = PROJECT_ROOT / "sdk" / "typescript"
    SDK_GEN_ROOT_DIR = PROJECT_ROOT / "sdk" / "typescript-gen"
    GENERATOR_CONFIG_DIR = PROJECT_ROOT / ".openapi-generator" / "typescript"

    def __init__(self, version: str, target: RunMode | None = None) -> None:
        self.version = version
        self.target = target

    def generate(self) -> None:
        _cleanup_stale_sdk_root_files(self.SDK_ROOT_DIR)
        typescript_settings = settings.sdk.typescript

        audience_packages = {
            RunMode.PORTAL: {
                "dst_dir": self.SDK_ROOT_DIR / "portal",
                "gen_dir": self.SDK_GEN_ROOT_DIR / "portal",
                "config_overrides": {
                    "npmName": typescript_settings.portal.npm_name,
                    "npmDescription": typescript_settings.portal.npm_description,
                },
            },
            RunMode.ADMIN: {
                "dst_dir": self.SDK_ROOT_DIR / "admin",
                "gen_dir": self.SDK_GEN_ROOT_DIR / "admin",
                "config_overrides": {
                    "npmName": typescript_settings.admin.npm_name,
                    "npmDescription": typescript_settings.admin.npm_description,
                },
            },
        }

        target_modes = (
            [self.target]
            if self.target is not None
            else [RunMode.PORTAL, RunMode.ADMIN]
        )

        for mode in target_modes:
            spec = audience_packages[mode]
            _generate_typescript_helper(
                mode,
                self.version,
                spec["dst_dir"],
                spec["gen_dir"],
                self.GENERATOR_CONFIG_DIR,
                config_overrides=spec["config_overrides"],
            )


def _cleanup_stale_sdk_root_files(root_dir: Path) -> None:
    """Remove stale flat sdk/typescript files while keeping portal/admin packages."""
    if not root_dir.exists() or not root_dir.is_dir():
        return

    # Skip when the root only acts as a parent directory for portal/admin packages.
    if not (root_dir / "package.json").exists():
        return

    for child in root_dir.iterdir():
        if child.name in {"portal", "admin"}:
            continue
        if child.is_dir():
            shutil.rmtree(child)
            continue
        child.unlink(missing_ok=True)


def _generate_typescript_helper(
    mode: RunMode,
    version: str,
    dst_dir: Path,
    gen_dir: Path,
    generator_config_dir: Path,
    config_overrides: dict[str, str] | None = None,
) -> None:
    """
    Helper function to generate one TypeScript SDK package.
    1. Updates the generator config with the specified version.
    2. Generates the SDK using OpenAPI Generator in a Docker container.
    3. Post-processes the generated SDK.
    4. Cleans up temporary directories.
    """
    if mode not in {RunMode.PORTAL, RunMode.ADMIN}:
        raise ValueError(f"Invalid mode: {mode}. Must be 'portal' or 'admin'.")

    if mode == RunMode.PORTAL:
        msg = "Portal SDK"
    else:
        msg = "Admin SDK"

    # 1. Update generator config with the specified version
    generator_config = generator_config_dir / "config.json"
    with open(generator_config) as f:
        config_data = json.load(f)

    config_data["npmVersion"] = version
    if config_overrides:
        config_data.update(config_overrides)

    tmp_generator_config = generator_config_dir / "config_tmp.json"
    with open(tmp_generator_config, "w") as f:
        json.dump(config_data, f, indent=2)

    console.print(f"[bold green]✅ Updated npmVersion to {version}[/bold green]")

    # 2. Generate using OpenAPI Generator
    console.print(f"[bold blue]Step 1: Generating TypeScript {msg}..[/bold blue]")

    if gen_dir.exists():
        shutil.rmtree(gen_dir)

    gen_dir.mkdir(parents=True)

    volume_path = Path(settings.openapi.generator_volume_root)
    relative_swagger = SWAGGER_ROOT.relative_to(PROJECT_ROOT)
    relative_gen = gen_dir.relative_to(PROJECT_ROOT)
    relative_generator_config = generator_config_dir.relative_to(PROJECT_ROOT)

    container_input_path = (
        volume_path / relative_swagger / "converted" / f"{mode.value}.json"
    )
    container_output_path = volume_path / relative_gen
    container_config_path = volume_path / relative_generator_config / "config_tmp.json"
    container_templates_path = volume_path / relative_generator_config / "templates"

    # Get current user UID and GID to avoid permission issues
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
                "typescript-axios",
                "-o",
                container_output_path.as_posix(),
                "-c",
                container_config_path.as_posix(),
                "-t",
                container_templates_path.as_posix(),
            ],
            volumes=[(PROJECT_ROOT, volume_path)],
            user=f"{current_user}:{current_group}",
            remove=True,
        )
    except Exception as e:
        console.print(
            f"[bold_red]❌ Error during typescript {msg} generation: {e}[/bold_red]"
        )
        sys.exit(1)
    finally:
        if tmp_generator_config.exists():
            tmp_generator_config.unlink(missing_ok=True)

    console.print(
        f"[bold green]✅ Generated TypeScript {msg} successfully![/bold green]"
    )
    console.print()

    # 3. Post-process generated SDK
    console.print(f"[bold blue]Step 2: Post-processing generated {msg}...[/bold blue]")

    # Clean up existing
    if dst_dir.exists():
        shutil.rmtree(dst_dir)

    # Copy the generated
    shutil.copytree(gen_dir, dst_dir)

    console.print(
        f"[bold green]✅ TypeScript {msg} post-processing completed successfully![/bold green]"
    )
    console.print()

    # 4. Clean up temporary directory
    if gen_dir.exists():
        shutil.rmtree(gen_dir)

    # 5. Build the TypeScript SDK
    console.print(f"[bold blue]Step 3: Building TypeScript {msg}...[/bold blue]")

    # Check if pnpm is available, fallback to npm
    pkg_manager = "pnpm" if shutil.which("pnpm") else "npm"
    console.print(f"[dim]Using package manager: {pkg_manager}[/dim]")

    run_command([pkg_manager, "install"], cwd=dst_dir, capture_output=True, text=True)
    console.print("[dim]✓ Dependencies installed[/dim]")

    run_command(
        [pkg_manager, "run", "build"], cwd=dst_dir, capture_output=True, text=True
    )
    console.print("[dim]✓ TypeScript compiled[/dim]")

    console.print(f"[bold green]✅ TypeScript {msg} built successfully![/bold green]")
