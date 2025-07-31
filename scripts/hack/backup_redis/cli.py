#!/usr/bin/env -S uv run -s
from typing import Optional
from psycopg2.extras import RealDictCursor
from rich.console import Console
from redis import Redis
import psycopg2
import redis
import typer

KEY_FORMAT = "trace:{}:log"

DEFAULT_LOCAL_URL = "redis://localhost:6379"
DEFAULT_PG_HOST = "10.10.10.220"
DEFAULT_PG_PORT = "32432"
DEFAULT_PG_USER = "postgres"
DEFAULT_PG_PASSWORD = "yourpassword"
DEFAULT_PG_DB = "rcabench"
DEFAULT_REMOTE_URL = "redis://10.10.10.220:32279"
DEFAULT_REDIS_DB = 0


app = typer.Typer(help="Redis 备份工具")
console = Console()


class Client:
    def __init__(self, source_url: str, target_url: str) -> None:
        self.source_redis, self.target_redis = self._connect_redis(
            source_url, target_url
        )
        self.pg_client = self._connect_postgres()

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

    def _connect_postgres(self):
        """连接到 PostgreSQL 数据库"""
        try:
            console.print(
                f"[cyan]连接 PostgreSQL: {DEFAULT_PG_HOST}:{DEFAULT_PG_PORT}/{DEFAULT_PG_DB}[/cyan]"
            )

            conn = psycopg2.connect(
                host=DEFAULT_PG_HOST,
                port=DEFAULT_PG_PORT,
                user=DEFAULT_PG_USER,
                password=DEFAULT_PG_PASSWORD,
                database=DEFAULT_PG_DB,
                cursor_factory=RealDictCursor,
            )

            # 测试连接
            with conn.cursor() as cursor:
                cursor.execute("SELECT version();")
                version = cursor.fetchone()
                console.print("[green]✅ PostgreSQL 连接成功[/green]")
                console.print(f"[dim]版本: {version['version'][:50]}...[/dim]")  # type: ignore

            return conn

        except psycopg2.Error as e:
            console.print(f"[red]PostgreSQL 连接失败: {e}[/red]")
            raise typer.Exit(code=1)
        except Exception as e:
            console.print(f"[red]未知错误: {e}[/red]")
            raise typer.Exit(code=1)

    def _read_hashes_fuzzy(self, pattern: str) -> list[str]:
        console.print(f"[cyan]查找匹配 '{pattern}' 的哈希表...[/cyan]")

        matching_keys = self.source_redis.keys(pattern)
        if not matching_keys:
            console.print(f"[red]没有找到匹配 '{pattern}' 的键[/red]")
            raise typer.Exit(code=1)

        hashes = []
        for key in matching_keys:  # type: ignore
            if self.source_redis.type(key) == "hash":
                hashes.append(key)

        if not hashes:
            console.print("[yellow]匹配的键中没有哈希表类型的数据[/yellow]")
            raise typer.Exit(code=1)

        console.print(f"[bold blue]找到 {len(hashes)} 个匹配的哈希表[/bold blue]")
        return hashes

    def _read_streams_exact(self) -> list[str]:
        console.print("[cyan]执行查询获取流名称...[/cyan]")

        query = """
        SELECT DISTINCT t.trace_id
        FROM fault_injection_schedules fis
        INNER JOIN tasks t ON fis.task_id = t.id
        WHERE fis.task_id IS NOT NULL;
        """

        try:
            console.print(f"[dim]SQL: {query}[/dim]")
            with self.pg_client.cursor() as cursor:
                cursor.execute(query)
                results = cursor.fetchall()

                if not results:
                    console.print("[yellow]查询结果为空[/yellow]")
                    return []

                streams = []
                for row in results:
                    if isinstance(row, dict):
                        trace_id = list(row.values())[0]
                    else:
                        trace_id = row[0]

                    if trace_id:
                        streams.append(KEY_FORMAT.format(trace_id))

                console.print(
                    f"[green]从 PostgreSQL 查询到 {len(streams)} 个流名称[/green]"
                )
                return streams

        except psycopg2.Error as e:
            console.print(f"[red]PostgreSQL 查询失败: {e}[/red]")
            raise typer.Exit(code=1)
        except Exception as e:
            console.print(f"[red]查询执行错误: {e}[/red]")
            raise typer.Exit(code=1)

    def _read_streams_fuzzy(self, pattern: str) -> list[str]:
        console.print(f"[cyan]查找匹配 '{pattern}' 的流...[/cyan]")

        matching_keys = self.source_redis.keys(pattern)
        if not matching_keys:
            console.print(f"[red]没有找到匹配 '{pattern}' 的流[/red]")
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

    def copy_hashes(
        self, pattern: str, overwrite: bool = False, dry_run: bool = False
    ) -> None:
        hashes = self._read_hashes_fuzzy(pattern)

        if dry_run:
            console.print("[yellow]Dry run 模式，不会实际复制数据[/yellow]")
            return

        console.print("[cyan]开始批量复制...[/cyan]")

        success_count = 0
        failed_count = 0
        skipped_count = 0

        for i, source_key in enumerate(hashes, 1):
            try:
                target_key = source_key

                # 检查源哈希表
                if not self.source_redis.exists(source_key):
                    console.print(
                        f"[yellow][{i}/{len(hashes)}] 跳过不存在: {source_key}[/yellow]"
                    )
                    skipped_count += 1
                    continue

                hash_length = self.source_redis.hlen(source_key)
                if hash_length == 0:
                    console.print(
                        f"[yellow][{i}/{len(hashes)}] 跳过空哈希表: {source_key}[/yellow]"
                    )
                    skipped_count += 1
                    continue

                # 检查目标哈希表
                target_exists = self.target_redis.exists(target_key)
                if target_exists and not overwrite:
                    target_length = self.target_redis.hlen(target_key)
                    console.print(
                        f"[yellow][{i}/{len(hashes)}] 跳过已存在: {source_key} ({target_length} 个字段)[/yellow]"
                    )
                    skipped_count += 1
                    continue

                # 显示当前处理的哈希表
                console.print(
                    f"[cyan][{i}/{len(hashes)}] 复制: {source_key} ({hash_length} 个字段)[/cyan]"
                )

                # 如果需要覆盖，先删除目标哈希表
                if target_exists and overwrite:
                    self.target_redis.delete(target_key)

                # 获取并复制哈希数据
                all_fields = self.source_redis.hgetall(source_key)
                if not all_fields:
                    skipped_count += 1
                    continue

                try:
                    # 批量设置哈希字段
                    self.target_redis.hset(target_key, mapping=all_fields)  # type: ignore
                    success_count += 1
                    console.print(
                        f"[dim]  [green]✓[/green] {len(all_fields)} 个字段[/dim]"  # type: ignore
                    )
                except Exception as e:
                    failed_count += 1
                    console.print(f"[red]  ✗ 复制失败: {e}[/red]")

                # 每10个哈希表显示一次进度
                if i % 10 == 0:
                    console.print(
                        f"[dim]进度: {i}/{len(hashes)} ({success_count} 成功, {failed_count} 失败, {skipped_count} 跳过)[/dim]"
                    )

            except Exception as e:
                failed_count += 1
                console.print(
                    f"[red][{i}/{len(hashes)}] 处理失败: {source_key} - {e}[/red]"
                )

        # 显示最终结果
        console.print("\n[bold]批量复制完成[/bold]")
        console.print(f"[green]✅ 成功: {success_count} 个哈希表[/green]")

        if failed_count > 0:
            console.print(f"[red]❌ 失败: {failed_count} 个哈希表[/red]")

        if skipped_count > 0:
            console.print(f"[yellow]🚫 跳过: {skipped_count} 个哈希表[/yellow]")

        console.print(f"[blue]总计处理: {len(hashes)} 个哈希表[/blue]")

    def copy_streams(
        self,
        exact_match: bool = False,
        fuzzy_match: str | None = None,
        overwrite: bool = False,
        dry_run: bool = False,
    ) -> None:
        if exact_match:
            streams = self._read_streams_exact()

        if fuzzy_match is not None:
            streams = self._read_streams_fuzzy(fuzzy_match)

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
def restore_hashes(
    source_url: str = typer.Option(
        DEFAULT_REMOTE_URL, "--source_url", help="源 Redis URL"
    ),
    target_url: str = typer.Option(
        DEFAULT_LOCAL_URL, "--target_url", help="目标 Redis URL"
    ),
    pattern: str = typer.Option(
        "trace:*:log", "--pattern", help="源 Redis 哈希表匹配模式"
    ),
    overwrite: bool = typer.Option(
        False, "--overwrite", help="是否覆盖目标哈希表（默认：False）"
    ),
    dry_run: bool = typer.Option(False, "--dry_run", help="是否只显示操作而不实际执行"),
):
    """批量恢复 Redis 哈希表数据

    支持模糊匹配模式，使用 Redis KEYS 命令查找匹配指定模式的哈希表。

    """

    client = Client(source_url, target_url)
    client.copy_hashes(pattern, overwrite, dry_run)


@app.command()
def restore_streams(
    source_url: str = typer.Option(
        DEFAULT_REMOTE_URL, "--source_url", help="源 Redis URL"
    ),
    target_url: str = typer.Option(
        DEFAULT_LOCAL_URL, "--target_url", help="目标 Redis URL"
    ),
    exact_match: bool = typer.Option(
        False, "--exact_match", help="源 Redis 流精确匹配（默认：False）"
    ),
    fuzzy_match: Optional[str] = typer.Option(
        None, "--fuzzy_match", help="源 Redis 流模糊匹配"
    ),
    overwrite: bool = typer.Option(
        False, "--overwrite", help="是否覆盖目标流（默认：True）"
    ),
    dry_run: bool = typer.Option(False, "--dry_run", help="是否只显示操作而不实际执行"),
):
    """批量恢复 Redis Stream 数据

    支持两种匹配模式：
    1. 精确匹配：从 PostgreSQL 数据库查询 fault_injection_schedules 表获取对应的 trace_id
    2. 模糊匹配：使用 Redis KEYS 命令查找匹配指定模式的流

    """

    # 参数验证：必须提供其中一个，且不能同时提供
    if not exact_match and not fuzzy_match:
        console.print("[red]错误：必须选择一种匹配模式[/red]")
        console.print(
            "[yellow]使用 --exact_match 从数据库精确匹配，或使用 --fuzzy_match 进行模糊匹配[/yellow]"
        )
        raise typer.Exit(code=1)

    if exact_match and fuzzy_match:
        console.print("[red]错误：不能同时使用两种匹配模式[/red]")
        console.print("[yellow]请选择 --exact_match 或 --fuzzy_match 其中一个[/yellow]")
        raise typer.Exit(code=1)

    client = Client(source_url, target_url)
    client.copy_streams(exact_match, fuzzy_match, overwrite, dry_run)


if __name__ == "__main__":
    app()
