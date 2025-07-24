#!/usr/bin/env -S uv run -s
from typing import Optional
from pathlib import Path
from rich.console import Console
from redis import Redis
import redis
import typer

DEFAULT_LOCAL_URL = "redis://localhost:6379"
DEFAULT_REMOTE_URL = "redis://10.10.10.220:32279"
DEFAULT_REDIS_DB = 0

app = typer.Typer(help="Redis 备份工具")
console = Console()


class Client:
    def __init__(self, source_url: str, target_url: str) -> None:
        self.source_redis, self.target_redis = self._connect_redis(
            source_url, target_url
        )

    def _connect_redis(self, source_url: str, target_url: str) -> tuple[Redis, Redis]:
        """连接到源和目标 Redis，返回连接对象"""
        console.print(f"[cyan]连接源 Redis: {source_url}[/cyan]")
        console.print(f"[cyan]连接目标 Redis: {target_url}[/cyan]")

        try:
            source_redis: Redis = redis.from_url(source_url, decode_responses=True)
            target_redis: Redis = redis.from_url(target_url, decode_responses=True)

            console.print("[cyan]测试源 Redis 连接...[/cyan]")
            source_redis.ping()
            console.print("[green]✅ 源 Redis 连接成功[/green]")

            console.print("[cyan]测试目标 Redis 连接...[/cyan]")
            target_redis.ping()
            console.print("[green]✅ 目标 Redis 连接成功[/green]")

            return source_redis, target_redis

        except redis.ConnectionError as e:
            console.print(f"[red]Redis 连接失败: {e}[/red]")
            raise typer.Exit(code=1)
        except Exception as e:
            console.print(f"[red]未知错误: {e}[/red]")
            raise typer.Exit(code=1)

    def _read_streams_from_file(self, file_path: str) -> list[str]:
        """从文件读取流名称列表"""
        try:
            path = Path(file_path)
            if not path.exists():
                console.print(f"[red]文件不存在: {file_path}[/red]")
                raise typer.Exit(code=1)

            console.print(f"[cyan]从文件读取流列表: {file_path}[/cyan]")

            with open(path, "r", encoding="utf-8") as f:
                lines = f.readlines()

            # 处理文件内容
            streams = []
            for line in lines:
                line = line.strip()

                # 跳过空行和注释行
                if not line or line.startswith("#"):
                    continue

                # 支持多种分隔符
                if "," in line:
                    # CSV 格式：stream1,stream2,stream3
                    names = [name.strip() for name in line.split(",") if name.strip()]
                    streams.extend(names)
                else:
                    # 每行一个流名称
                    streams.append(line)

            if not streams:
                console.print(
                    f"[yellow]文件 {file_path} 中没有找到有效的流名称[/yellow]"
                )
                raise typer.Exit(code=1)

            console.print(f"[green]从文件读取到 {len(streams)} 个流名称[/green]")
            return streams

        except Exception as e:
            console.print(f"[red]读取文件失败: {e}[/red]")
            raise typer.Exit(code=1)

    def _read_streams_from_pattern(self, source_pattern: str) -> list[str]:
        console.print(f"[cyan]查找匹配 '{source_pattern}' 的流...[/cyan]")

        matching_keys = self.source_redis.keys(source_pattern)
        if not matching_keys:
            console.print(f"[red]没有找到匹配 '{source_pattern}' 的流[/red]")
            raise typer.Exit(code=1)

        # 过滤出真正的 stream 类型
        streams = []
        for key in matching_keys:  # type: ignore
            if self.source_redis.type(key) == "stream":
                streams.append(key)

        if not streams:
            console.print("[yellow]匹配的键中没有 stream 类型的数据[/yellow]")
            raise typer.Exit(code=1)

        console.print(f"[bold blue]找到 {len(streams)} 个匹配的流:[/bold blue]")

        return streams

    def copy_stream_one(
        self, source_key: str, overwrite: bool = True, dry_run: bool = False
    ) -> None:
        target_key = source_key

        if not self.source_redis.exists(source_key):
            console.print(f"[red]源追踪流 {source_key} 不存在[/red]")
            raise typer.Exit(code=1)

        stream_length = self.source_redis.xlen(source_key)
        if stream_length == 0:
            console.print(f"[yellow]源流 {source_key} 为空[/yellow]")
            raise typer.Exit(code=1)

        console.print(f"[cyan]源流包含 {stream_length} 条记录[/cyan]")

        # 检查目标流是否已存在
        target_exists = self.target_redis.exists(target_key)
        if target_exists:
            target_length = self.target_redis.xlen(target_key)
            console.print(f"[yellow]目标流已存在，包含 {target_length} 条记录[/yellow]")

            # 询问是否覆盖
            if not overwrite:
                raise typer.Exit(code=0)

        # 获取所有追踪事件
        messages = self.source_redis.xrange(source_key)
        if not messages:
            console.print("[yellow]源追踪流为空[/yellow]")
            raise typer.Exit(code=1)

        console.print(f"[cyan]开始复制 {len(messages)} 条记录...[/cyan]")  # type: ignore

        copied_count, failed_count = 0, 0

        if dry_run:
            console.print("[yellow]Dry run 模式，不会实际复制数据[/yellow]")
            return

        for msg_id, fields in messages:  # type: ignore
            try:
                # 复制事件数据
                self.target_redis.xadd(target_key, fields)
                copied_count += 1

                # 每100条记录显示一次进度
                if copied_count % 100 == 0:
                    console.print(f"[dim]已复制 {copied_count} 条记录...[/dim]")

            except Exception as e:
                failed_count += 1
                console.print(f"[red]复制事件 {msg_id} 失败: {e}[/red]")
                console.print(f"[red]失败的字段: {fields}[/red]")

        console.print("[bold]复制完成:[/bold]")
        console.print(f"[green]✅ 成功: {copied_count} 条记录[/green]")
        if failed_count > 0:
            console.print(f"[red]❌ 失败: {failed_count} 条记录[/red]")

    def copy_stream_batch(
        self,
        file_path: str | None = None,
        source_pattern: str | None = None,
        overwrite: bool = True,
        dry_run: bool = False,
    ) -> None:
        if file_path is not None:
            streams = self._read_streams_from_file(file_path)

        if source_pattern is not None:
            streams = self._read_streams_from_pattern(source_pattern)

        if dry_run:
            console.print("[yellow]Dry run 模式，不会实际复制数据[/yellow]")
            return

        console.print("[cyan]开始批量复制...[/cyan]")

        success_count = 0
        failed_count = 0
        skipped_count = 0

        for i, source_key in enumerate(streams, 1):
            try:
                target_key = source_key

                # 检查源流
                if not self.source_redis.exists(source_key):
                    console.print(
                        f"[yellow][{i}/{len(streams)}] 跳过不存在: {source_key}[/yellow]"
                    )
                    skipped_count += 1
                    continue

                stream_length = self.source_redis.xlen(source_key)
                if stream_length == 0:
                    console.print(
                        f"[yellow][{i}/{len(streams)}] 跳过空流: {source_key}[/yellow]"
                    )
                    skipped_count += 1
                    continue

                # 检查目标流
                target_exists = self.target_redis.exists(target_key)
                if target_exists and not overwrite:
                    target_length = self.target_redis.xlen(target_key)
                    console.print(
                        f"[yellow][{i}/{len(streams)}] 跳过已存在: {source_key} ({target_length} 条记录)[/yellow]"
                    )
                    skipped_count += 1
                    continue

                # 显示当前处理的流
                console.print(
                    f"[cyan][{i}/{len(streams)}] 复制: {source_key} ({stream_length} 条记录)[/cyan]"
                )

                # 如果需要覆盖，先删除目标流
                if target_exists and overwrite:
                    self.target_redis.delete(target_key)

                # 获取并复制消息
                messages = self.source_redis.xrange(source_key)
                if not messages:
                    skipped_count += 1
                    continue

                copied_count = 0
                message_failed_count = 0

                for _, fields in messages:  # type: ignore
                    try:
                        self.target_redis.xadd(target_key, fields)
                        copied_count += 1
                    except Exception:
                        message_failed_count += 1

                if copied_count > 0:
                    success_count += 1
                    status_msg = "[green]✓[/green]"
                    if message_failed_count > 0:
                        status_msg += f" [yellow]({message_failed_count} 失败)[/yellow]"

                    console.print(f"[dim]  {status_msg} {copied_count} 条记录[/dim]")
                else:
                    failed_count += 1
                    console.print("[red]  ✗ 复制失败[/red]")

                # 每10个流显示一次进度
                if i % 10 == 0:
                    console.print(
                        f"[dim]进度: {i}/{len(streams)} ({success_count} 成功, {failed_count} 失败, {skipped_count} 跳过)[/dim]"
                    )

            except Exception as e:
                failed_count += 1
                console.print(
                    f"[red][{i}/{len(streams)}] 处理失败: {source_key} - {e}[/red]"
                )

        # 显示最终结果
        console.print("\n[bold]批量复制完成[/bold]")
        console.print(f"[green]✅ 成功: {success_count} 个流[/green]")

        if failed_count > 0:
            console.print(f"[red]❌ 失败: {failed_count} 个流[/red]")

        if skipped_count > 0:
            console.print(f"[yellow]🚫 跳过: {skipped_count} 个流[/yellow]")

        console.print(f"[blue]总计处理: {len(streams)} 个流[/blue]")


@app.command()
def restore_one(
    source_url: str = typer.Option(
        DEFAULT_REMOTE_URL, "--source_url", help="源 Redis URL"
    ),
    target_url: str = typer.Option(
        DEFAULT_LOCAL_URL, "--target_url", help="目标 Redis URL"
    ),
    source_key: Optional[str] = typer.Option(None, "--source_key", help="源 Redis 流"),
    overwrite: bool = typer.Option(
        True, "--overwrite", help="是否覆盖目标流（默认：True）"
    ),
    dry_run: bool = typer.Option(False, "--dry_run", help="是否只显示操作而不实际执行"),
) -> None:
    """备份指定名称的 Stream 数据"""
    if not source_key:
        console.print("[red]请提供源 Redis 流名称 (--source_key)[/red]")
        raise typer.Exit(code=1)

    client = Client(source_url, target_url)
    client.copy_stream_one(source_key, overwrite, dry_run)


@app.command()
def restore_batch(
    source_url: str = typer.Option(
        DEFAULT_REMOTE_URL, "--source_url", help="源 Redis URL"
    ),
    target_url: str = typer.Option(
        DEFAULT_LOCAL_URL, "--target_url", help="目标 Redis URL"
    ),
    file_path: Optional[str] = typer.Option(
        None, "--file_path", help="包含流名称的文件路径"
    ),
    source_pattern: Optional[str] = typer.Option(
        None, "--source_pattern", help="源 Redis 流匹配模式"
    ),
    overwrite: bool = typer.Option(
        True, "--overwrite", help="是否覆盖目标流（默认：True）"
    ),
    dry_run: bool = typer.Option(False, "--dry_run", help="是否只显示操作而不实际执行"),
):
    """批量恢复 Redis Stream 数据

    支持两种模式：
    1. 文件模式：使用 --file_path 指定包含流名称的文件
    2. 模式匹配：使用 --source_pattern 指定匹配模式

    示例：
    文件模式：uv run python cli.py restore-batch --file_path streams.txt
    模式匹配：uv run python cli.py restore-batch --source_pattern "trace:*:log"
    """

    # 参数验证：必须提供其中一个，且不能同时提供
    if not file_path and not source_pattern:
        console.print("[red]请提供以下参数之一:[/red]")
        console.print("[red]  --file_path: 包含流名称的文件路径[/red]")
        console.print("[red]  --source_pattern: Redis 流匹配模式[/red]")
        raise typer.Exit(code=1)

    if file_path and source_pattern:
        console.print("[red]请只提供一个参数，不能同时使用:[/red]")
        console.print("[red]  --file_path 和 --source_pattern[/red]")
        raise typer.Exit(code=1)

    client = Client(source_url, target_url)
    client.copy_stream_batch(file_path, source_pattern, overwrite, dry_run)


if __name__ == "__main__":
    app()
