#!/usr/bin/env -S uv run -s
from pathlib import Path
from rich.console import Console
import redis
import typer

BACKUP_DIR = Path("./temp/backup_redis")
BACKUP_DIR.mkdir(parents=True, exist_ok=True)

DEFAULT_REDIS_HOST = "10.10.10.220"
DEFAULT_REDIS_PORT = "32279"
DEFAULT_REDIS_DB = 0

app = typer.Typer(help="Redis 备份工具")
console = Console()


@app.command()
def backup_one(
    source_url: str = typer.Option(
        f"redis://{DEFAULT_REDIS_HOST}:{DEFAULT_REDIS_PORT}",
        "--source_url",
        help="源 Redis URL",
    ),
    target_url: str = typer.Option(
        "redis://localhost:6379", "--target_url", help="目标 Redis URL"
    ),
    source_key: str = typer.Option(None, "--source_key", help="源 Redis 流"),
    target_key: str = typer.Option(None, "--target_key", help="目标 Redis 流"),
) -> None:
    """备份指定名称的 Stream 数据"""
    source_redis = redis.from_url(source_url, decode_responses=True)
    target_redis = redis.from_url(target_url, decode_responses=True)

    if not source_key:
        console.print("[red]请提供源 Redis 流名称 (--source_key)[/red]")
        raise typer.Exit(code=1)

    if not target_key:
        target_key = source_key

    console.print(f"[bold blue]复制追踪流: {source_key} -> {target_key}[/bold blue]")

    # 检查源流是否存在
    if not source_redis.exists(source_key):
        console.print(f"[red]源追踪流 {source_key} 不存在[/red]")
        raise typer.Exit(code=1)

    # 获取所有追踪事件
    messages = source_redis.xrange(source_key)

    if not messages:
        console.print("[yellow]源追踪流为空[yellow]")
        raise typer.Exit(code=1)

    copied_count = 0
    for msg_id, fields in messages:  # type: ignore
        try:
            # 复制事件数据
            target_redis.xadd(target_key, fields)
            copied_count += 1
            console.print(f"[dim]复制事件 {msg_id} [/dim]")

        except Exception as e:
            console.print(f"[red]复制事件 {msg_id} 失败: {e}[/red]")

    console.print(f"[green]追踪流复制完成，共复制 {copied_count} 条事件[/green]")


if __name__ == "__main__":
    app()
