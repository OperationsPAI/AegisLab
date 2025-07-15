#!/usr/bin/env -S uv run -s
from rich.console import Console
import redis
import typer


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
    console.print(f"[cyan]连接源 Redis: {source_url}[/cyan]")
    console.print(f"[cyan]连接目标 Redis: {target_url}[/cyan]")

    try:
        source_redis = redis.from_url(source_url, decode_responses=True)
        target_redis = redis.from_url(target_url, decode_responses=True)

        # 测试连接
        console.print("[cyan]测试源 Redis 连接...[/cyan]")
        source_redis.ping()
        console.print("[green]✓ 源 Redis 连接成功[/green]")

        console.print("[cyan]测试目标 Redis 连接...[/cyan]")
        target_redis.ping()
        console.print("[green]✓ 目标 Redis 连接成功[/green]")

    except Exception as e:
        console.print(f"[red]Redis 连接失败: {e}[/red]")
        raise typer.Exit(code=1)

    if not source_key:
        console.print("[red]请提供源 Redis 流名称 (--source_key)[/red]")
        raise typer.Exit(code=1)

    if not target_key:
        target_key = source_key

    console.print(f"[bold blue]复制追踪流: {source_key} -> {target_key}[/bold blue]")

    # 检查源流是否存在
    if not source_redis.exists(source_key):
        console.print(f"[red]源追踪流 {source_key} 不存在[/red]")

        console.print("[cyan]查找可用的流...[/cyan]")
        available_keys = source_redis.keys("*")
        if available_keys:
            console.print("[cyan]可用的键:[/cyan]")
            for key in available_keys[:10]:  # type: ignore # 只显示前10个
                console.print(f"  - {key}")
        else:
            console.print("[yellow]没有找到任何键[/yellow]")

        raise typer.Exit(code=1)

    stream_length = source_redis.xlen(source_key)
    console.print(f"[cyan]源流包含 {stream_length} 条记录[/cyan]")

    # 检查目标流是否已存在
    target_exists = target_redis.exists(target_key)
    if target_exists:
        target_length = target_redis.xlen(target_key)
        console.print(f"[yellow]目标流已存在，包含 {target_length} 条记录[/yellow]")

        # 询问是否覆盖
        if not typer.confirm("是否继续添加到现有流？"):
            raise typer.Exit(code=0)

    # 获取所有追踪事件
    messages = source_redis.xrange(source_key)

    if not messages:
        console.print("[yellow]源追踪流为空[/yellow]")
        raise typer.Exit(code=1)

    console.print(f"[cyan]开始复制 {len(messages)} 条记录...[/cyan]")  # type: ignore

    copied_count = 0
    failed_count = 0

    for msg_id, fields in messages:  # type: ignore
        try:
            # 复制事件数据
            target_redis.xadd(target_key, fields)
            copied_count += 1

            # 每100条记录显示一次进度
            if copied_count % 100 == 0:
                console.print(f"[dim]已复制 {copied_count} 条记录...[/dim]")

        except Exception as e:
            failed_count += 1
            console.print(f"[red]复制事件 {msg_id} 失败: {e}[/red]")

            # 显示失败的字段内容
            console.print(f"[red]失败的字段: {fields}[/red]")

    console.print(f"[green]追踪流复制完成，共复制 {copied_count} 条事件[/green]")


if __name__ == "__main__":
    app()
