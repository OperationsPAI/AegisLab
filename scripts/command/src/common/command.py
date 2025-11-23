import shlex
import subprocess
import sys
from collections.abc import Iterable
from pathlib import Path
from subprocess import CalledProcessError, CompletedProcess

from src.common.common import console


def run_command(
    cmd_list: Iterable[str],
    cwd: Path = Path.cwd(),
    check: bool = True,
    capture_output: bool = False,
    **kwargs,
) -> CompletedProcess[str]:
    """Runs a shell command and handles errors."""
    try:
        console.print(f"${' '.join(shlex.quote(c) for c in cmd_list)}")
        cmd_str_list = list(cmd_list)
        result = subprocess.run(
            cmd_str_list,
            cwd=cwd,
            check=check,
            text=True,
            capture_output=capture_output,
            **kwargs,
        )
        return result
    except CalledProcessError as e:
        console.print(
            f"[bold red]❌ Command failed: {' '.join(cmd_str_list)}[/bold red]"
        )
        if e.stderr:
            print(e.stderr)
        sys.exit(1)
