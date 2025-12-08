import shlex
import subprocess
import sys
from collections.abc import Iterable
from pathlib import Path
from subprocess import CalledProcessError, CompletedProcess

from src.common.common import console

__all__ = ["run_command", "run_pipeline"]


def run_command(
    cmd_list: Iterable[str],
    cwd: Path = Path.cwd(),
    check: bool = True,
    capture_output: bool = False,
    cmd_print: bool = True,
    **kwargs,
) -> CompletedProcess[str]:
    """Runs a shell command and handles errors."""
    try:
        if cmd_print:
            console.print(f"${' '.join(shlex.quote(c) for c in cmd_list)}")

        cmd_str_list = list(cmd_list)
        result = subprocess.run(
            cmd_str_list,
            cwd=cwd,
            check=check,
            capture_output=capture_output,
            **kwargs,
        )
        return result
    except CalledProcessError as e:
        console.print(
            f"[bold red]❌ Command failed: {' '.join(cmd_str_list)}[/bold red]"
        )
        if e.stderr:
            console.print(e.stderr)
        sys.exit(1)


def run_pipeline(
    cmd1: Iterable[str], cmd2: Iterable[str], cwd: Path = Path.cwd()
) -> CompletedProcess[str]:
    """Runs two shell commands in a pipeline and handles errors."""
    try:
        console.print(
            f"${' '.join(shlex.quote(c) for c in cmd1)} | {' '.join(shlex.quote(c) for c in cmd2)}"
        )
        p1 = subprocess.Popen(
            list(cmd1),
            cwd=cwd,
            stdout=subprocess.PIPE,
            text=True,
        )
        p2 = subprocess.Popen(
            list(cmd2),
            cwd=cwd,
            stdin=p1.stdout,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )

        if p1.stdout is not None:
            p1.stdout.close()

        stdout, stderr = p2.communicate()
        if p2.returncode != 0:
            console.print(
                f"[bold red]❌ Pipeline command failed: {' '.join(cmd2)}[/bold red]"
            )
            if stderr:
                print(stderr)
            sys.exit(1)
        return CompletedProcess(list(cmd2), p2.returncode, stdout, stderr)

    except CalledProcessError as e:
        console.print(
            f"[bold red]❌ Command failed: {' '.join(cmd1)} or {' '.join(cmd2)}[/bold red]"
        )
        if e.stderr:
            print(e.stderr)
        sys.exit(1)
