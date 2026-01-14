import typer

from src.common.common import ENV, console
from src.port_manager import PortForwardManager

app = typer.Typer()


@app.command()
def start(
    env: ENV = typer.Option(ENV.PROD, "--env", "-e", help="Environment: prod or test"),
    namespace: str = typer.Option(
        "exp", "--namespace", "-n", help="Kubernetes namespace to forward"
    ),
    keep_alive: bool = typer.Option(
        False, "--keep-alive", "-k", help="Keep forwarding in foreground"
    ),
):
    if env == ENV.DEV:
        console.print(
            "[bold red]❌ Port forwarding is not supported for 'dev' environments.[/bold red]"
        )
        raise typer.Exit(code=1)

    if env == ENV.STAGING:
        console.print(
            "[bold yellow]ℹ️  Port forwarding is not needed for 'staging' environments.[/bold yellow]"
        )
        raise typer.Exit(code=0)

    manager = PortForwardManager(env, namespace)
    manager.start_forwarding()

    if keep_alive:
        console.print("\n[bold yellow]Press Ctrl+C to stop forwarding...[/bold yellow]")
        try:
            import time

            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            manager.stop_all_forwards()
            console.print("[bold green]✅ Forwarding stopped[/bold green]")
    else:
        console.print(
            "\n[bold green]✅ Port forwarding started in background[/bold green]"
        )
        console.print(
            "[bold yellow]🔔 Use 'pkill -f kubectl.*port-forward' to stop[/bold yellow]"
        )
