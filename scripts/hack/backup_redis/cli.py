#!/usr/bin/env -S uv run -s
from typing import Optional

import redis
import sqlalchemy
import typer
from redis import Redis
from rich.console import Console
from sqlalchemy import create_engine, text
from sqlalchemy.orm import Session, sessionmaker

KEY_FORMAT = "trace:{}:log"

DEFAULT_LOCAL_URL = "redis://localhost:6379"
DEFAULT_REMOTE_URL = "redis://10.10.10.220:32279"
DEFAULT_REDIS_DB = 0

DEFAULT_DB_HOST = "10.10.10.220"
DEFAULT_DB_PORT = "32206"
DEFAULT_DB_USER = "root"
DEFAULT_DB_PASSWORD = "yourpassword"
DEFAULT_DB_DB = "rcabench"

app = typer.Typer(help="Redis 备份工具")
console = Console()


class Client:
    """
    Redis backup client for handling data transfer between Redis instances.

    This class manages connections to source and target Redis instances,
    as well as database connections for exact matching queries.

    Attributes:
        source_redis: Connection to the source Redis instance
        target_redis: Connection to the target Redis instance
        db_client: Database session for executing queries
    """

    def __init__(self, source_url: str, target_url: str) -> None:
        """
        Initialize Redis backup client with source and target URLs.

        Args:
            source_url: URL of the source Redis instance
            target_url: URL of the target Redis instance

        Raises:
            typer.Exit: If connection to Redis or database fails
        """
        self.source_redis, self.target_redis = self._connect_redis(source_url, target_url)
        self.db_client = self._connect_database()

    def _connect_redis(self, source_url: str, target_url: str) -> tuple[Redis, Redis]:
        """
        Connect to source and target Redis instances.

        Args:
            source_url: URL of the source Redis instance
            target_url: URL of the target Redis instance

        Returns:
            Tuple containing source and target Redis connection objects

        Raises:
            typer.Exit: If connection fails
        """
        console.print(f"[cyan]Connecting to source Redis: {source_url}[/cyan]")
        console.print(f"[cyan]Connecting to target Redis: {target_url}[/cyan]")

        try:
            source_redis: Redis = redis.from_url(source_url, decode_responses=True)
            target_redis: Redis = redis.from_url(target_url, decode_responses=True)

            console.print("[cyan]Testing source Redis connection...[/cyan]")
            source_redis.ping()
            console.print("[green]✅ Source Redis connection successful[/green]")

            console.print("[cyan]Testing target Redis connection...[/cyan]")
            target_redis.ping()
            console.print("[green]✅ Target Redis connection successful[/green]")

            return source_redis, target_redis

        except redis.ConnectionError as e:
            console.print(f"[red]Redis connection failed: {e}[/red]")
            raise typer.Exit(code=1)
        except Exception as e:
            console.print(f"[red]Unknown error: {e}[/red]")
            raise typer.Exit(code=1)

    def _connect_database(self) -> Session:
        """
        Connect to the MySQL database.

        Returns:
            SQLAlchemy session object for database operations

        Raises:
            typer.Exit: If database connection fails
        """
        try:
            console.print(f"[cyan]Connecting to database: {DEFAULT_DB_HOST}:{DEFAULT_DB_PORT}/{DEFAULT_DB_DB}[/cyan]")

            db_url = f"mysql+pymysql://{DEFAULT_DB_USER}:{DEFAULT_DB_PASSWORD}@{DEFAULT_DB_HOST}:{DEFAULT_DB_PORT}/{DEFAULT_DB_DB}"
            engine = create_engine(db_url, echo=False, pool_pre_ping=True)

            Session = sessionmaker(bind=engine)
            session = Session()

            # Test connection
            result = session.execute(text("SELECT version()"))
            version = result.scalar()
            console.print("[green]✅ Database connection successful[/green]")
            console.print(f"[dim]Version: {version[:50]}...[/dim]")  # type: ignore

            return session

        except sqlalchemy.except_all.SQLAlchemyError as e:
            console.print(f"[red]❌ Database connection failed: {e}[/red]")
            raise typer.Exit(code=1)
        except Exception as e:
            console.print(f"[red]Unknown error: {e}[/red]")
            raise typer.Exit(code=1)

    def _read_hashes_fuzzy(self, pattern: str) -> list[str]:
        """
        Find hash tables matching the given pattern using fuzzy matching.

        Args:
            pattern: Redis key pattern to match (e.g., "trace:*:log")

        Returns:
            List of hash table keys that match the pattern

        Raises:
            typer.Exit: If no matching keys are found
        """
        console.print(f"[cyan]Searching for hash tables matching '{pattern}'...[/cyan]")

        matching_keys = self.source_redis.keys(pattern)
        if not matching_keys:
            console.print(f"[red]No keys found matching '{pattern}'[/red]")
            raise typer.Exit(code=1)

        hashes = []
        for key in matching_keys:  # type: ignore
            if self.source_redis.type(key) == "hash":
                hashes.append(key)

        if not hashes:
            console.print("[yellow]No hash table data found among matching keys[/yellow]")
            raise typer.Exit(code=1)

        console.print(f"[bold blue]Found {len(hashes)} matching hash tables[/bold blue]")
        return hashes

    def _read_streams_exact(self) -> list[str]:
        """
        Get stream names by querying the database for exact matches.

        Queries the fault_injection_schedules table to get trace_ids and
        constructs Redis stream keys from them.

        Returns:
            List of Redis stream keys

        Raises:
            typer.Exit: If database query fails
        """
        console.print("[cyan]Executing query to get stream names...[/cyan]")

        query = """
        SELECT DISTINCT t.trace_id
        FROM fault_injection_schedules fis
        INNER JOIN tasks t ON fis.task_id = t.id
        WHERE fis.task_id IS NOT NULL;
        """

        try:
            console.print(f"[dim]SQL: {query}[/dim]")

            result = self.db_client.execute(text(query))
            results = result.fetchall()
            if not results:
                console.print("[yellow]Query returned no results[/yellow]")
                return []

            streams = []
            for row in results:
                trace_id = row[0]  # SQLAlchemy result is in tuple format
                if trace_id:
                    streams.append(KEY_FORMAT.format(trace_id))

            console.print(f"[green]✅ Retrieved {len(streams)} stream names from database[/green]")
            return streams

        except sqlalchemy.except_all.SQLAlchemyError as e:
            console.print(f"[red]❌ Database query failed: {e}[/red]")
            raise typer.Exit(code=1)
        except Exception as e:
            console.print(f"[red]Query execution error: {e}[/red]")
            raise typer.Exit(code=1)

    def _read_streams_fuzzy(self, pattern: str) -> list[str]:
        """Find streams matching the given pattern using fuzzy matching.

        Args:
            pattern: Redis key pattern to match

        Returns:
            List of stream keys that match the pattern

        Raises:
            typer.Exit: If no matching streams are found
        """
        console.print(f"[cyan]Searching for streams matching '{pattern}'...[/cyan]")

        matching_keys = self.source_redis.keys(pattern)
        if not matching_keys:
            console.print(f"[red]No streams found matching '{pattern}'[/red]")
            raise typer.Exit(code=1)

        # Filter for actual stream type
        streams = []
        for key in matching_keys:  # type: ignore
            if self.source_redis.type(key) == "stream":
                streams.append(key)

        if not streams:
            console.print("[yellow]No stream type data found among matching keys[/yellow]")
            raise typer.Exit(code=1)

        console.print(f"[bold blue]Found {len(streams)} matching streams[/bold blue]")
        return streams

    def copy_hashes(self, pattern: str, overwrite: bool = False, dry_run: bool = False) -> None:
        """
        Copy hash tables from source to target Redis instance.

        Args:
            pattern: Redis key pattern to match hash tables
            overwrite: Whether to overwrite existing hash tables in target
            dry_run: If True, show operations without executing them

        This method finds hash tables matching the pattern, then copies each one
        from the source Redis to the target Redis. It provides detailed progress
        information and handles errors gracefully.
        """
        hashes = self._read_hashes_fuzzy(pattern)

        if dry_run:
            console.print("[yellow]Dry run mode, no data will be actually copied[/yellow]")
            return

        console.print("[cyan]Starting batch copy...[/cyan]")

        success_count = 0
        failed_count = 0
        skipped_count = 0

        for i, source_key in enumerate(hashes, 1):
            try:
                target_key = source_key

                # Check source hash table
                if not self.source_redis.exists(source_key):
                    console.print(f"[yellow][{i}/{len(hashes)}] Skipping non-existent: {source_key}[/yellow]")
                    skipped_count += 1
                    continue

                hash_length = self.source_redis.hlen(source_key)
                if hash_length == 0:
                    console.print(f"[yellow][{i}/{len(hashes)}] Skipping empty hash table: {source_key}[/yellow]")
                    skipped_count += 1
                    continue

                # Check target hash table
                target_exists = self.target_redis.exists(target_key)
                if target_exists and not overwrite:
                    target_length = self.target_redis.hlen(target_key)
                    console.print(
                        f"[yellow][{i}/{len(hashes)}] Skipping existing: {source_key} ({target_length} fields)[/yellow]"
                    )
                    skipped_count += 1
                    continue

                # Display current hash table being processed
                console.print(f"[cyan][{i}/{len(hashes)}] Copying: {source_key} ({hash_length} fields)[/cyan]")

                # Delete target hash table if overwrite is needed
                if target_exists and overwrite:
                    self.target_redis.delete(target_key)

                # Get and copy hash data
                all_fields = self.source_redis.hgetall(source_key)
                if not all_fields:
                    skipped_count += 1
                    continue

                try:
                    # Batch set hash fields
                    self.target_redis.hset(target_key, mapping=all_fields)  # type: ignore
                    success_count += 1
                    console.print(
                        f"[dim]  [green]✓[/green] {len(all_fields)} fields[/dim]"  # type: ignore
                    )
                except Exception as e:
                    failed_count += 1
                    console.print(f"[red]  ✗ Copy failed: {e}[/red]")

                # Show progress every 10 hash tables
                if i % 10 == 0:
                    console.print(
                        f"[dim]Progress: {i}/{len(hashes)} ({success_count} success, {failed_count} failed, {skipped_count} skipped)[/dim]"  # noqa: E501
                    )

            except Exception as e:
                failed_count += 1
                console.print(f"[red][{i}/{len(hashes)}] Processing failed: {source_key} - {e}[/red]")

        # Display final results
        console.print("\n[bold]Batch copy completed[/bold]")
        console.print(f"[green]✅ Success: {success_count} hash tables[/green]")

        if failed_count > 0:
            console.print(f"[red]❌ Failed: {failed_count} hash tables[/red]")

        if skipped_count > 0:
            console.print(f"[yellow]🚫 Skipped: {skipped_count} hash tables[/yellow]")

        console.print(f"[blue]Total processed: {len(hashes)} hash tables[/blue]")

    def copy_streams(
        self,
        exact_match: bool = False,
        fuzzy_match: str | None = None,
        overwrite: bool = False,
        dry_run: bool = False,
    ) -> None:
        """
        Copy Redis streams from source to target instance.

        Args:
            exact_match: Use database query for exact stream matching
            fuzzy_match: Pattern for fuzzy stream matching
            overwrite: Whether to overwrite existing streams in target
            dry_run: If True, show operations without executing them

        This method supports two modes of operation:
        1. Exact match: Query database to find specific trace IDs
        2. Fuzzy match: Use Redis KEYS command with pattern matching
        """
        if exact_match:
            streams = self._read_streams_exact()

        if fuzzy_match is not None:
            streams = self._read_streams_fuzzy(fuzzy_match)

        if dry_run:
            console.print("[yellow]Dry run mode, no data will be actually copied[/yellow]")
            return

        console.print("[cyan]Starting batch copy...[/cyan]")

        success_count = 0
        failed_count = 0
        skipped_count = 0

        for i, source_key in enumerate(streams, 1):
            try:
                target_key = source_key

                # Check source stream
                if not self.source_redis.exists(source_key):
                    console.print(f"[yellow][{i}/{len(streams)}] Skipping non-existent: {source_key}[/yellow]")
                    skipped_count += 1
                    continue

                stream_length = self.source_redis.xlen(source_key)
                if stream_length == 0:
                    console.print(f"[yellow][{i}/{len(streams)}] Skipping empty stream: {source_key}[/yellow]")
                    skipped_count += 1
                    continue

                # Check target stream
                target_exists = self.target_redis.exists(target_key)
                if target_exists and not overwrite:
                    target_length = self.target_redis.xlen(target_key)
                    console.print(
                        f"[yellow][{i}/{len(streams)}] Skipping existing: {source_key} ({target_length} records)[/yellow]"  # noqa: E501
                    )
                    skipped_count += 1
                    continue

                # Display current stream being processed
                console.print(f"[cyan][{i}/{len(streams)}] Copying: {source_key} ({stream_length} records)[/cyan]")

                # Delete target stream if overwrite is needed
                if target_exists and overwrite:
                    self.target_redis.delete(target_key)

                # Get and copy messages
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
                    status_msg = "[green]✅[/green]"
                    if message_failed_count > 0:
                        status_msg += f" [yellow]({message_failed_count} failed)[/yellow]"

                    console.print(f"[dim]  {status_msg} {copied_count} records[/dim]")
                else:
                    failed_count += 1
                    console.print("[red]❌ Copy failed[/red]")

                # Show progress every 10 streams
                if i % 10 == 0:
                    console.print(
                        f"[dim]Progress: {i}/{len(streams)} ({success_count} success, {failed_count} failed, {skipped_count} skipped)[/dim]"  # noqa: E501
                    )

            except Exception as e:
                failed_count += 1
                console.print(f"[red][{i}/{len(streams)}] Processing failed: {source_key} - {e}[/red]")

        # Display final results
        console.print("\n[bold]Batch copy completed[/bold]")
        console.print(f"[green]✅ Success: {success_count} streams[/green]")

        if failed_count > 0:
            console.print(f"[red]❌ Failed: {failed_count} streams[/red]")

        if skipped_count > 0:
            console.print(f"[yellow]🚫 Skipped: {skipped_count} streams[/yellow]")

        console.print(f"[blue]Total processed: {len(streams)} streams[/blue]")


@app.command()
def restore_hashes(
    source_url: str = typer.Option(DEFAULT_REMOTE_URL, "--source_url", help="Source Redis URL"),
    target_url: str = typer.Option(DEFAULT_LOCAL_URL, "--target_url", help="Target Redis URL"),
    pattern: str = typer.Option("trace:*:log", "--pattern", help="Source Redis hash table matching pattern"),
    overwrite: bool = typer.Option(
        False, "--overwrite", help="Whether to overwrite target hash tables (default: False)"
    ),
    dry_run: bool = typer.Option(False, "--dry_run", help="Show operations without actually executing them"),
):
    """
    Batch restore Redis hash table data.

    Supports fuzzy matching mode, using Redis KEYS command to find hash tables
    matching the specified pattern.
    """
    client = Client(source_url, target_url)
    client.copy_hashes(pattern, overwrite, dry_run)


@app.command()
def restore_streams(
    source_url: str = typer.Option(DEFAULT_REMOTE_URL, "--source_url", help="Source Redis URL"),
    target_url: str = typer.Option(DEFAULT_LOCAL_URL, "--target_url", help="Target Redis URL"),
    exact_match: bool = typer.Option(
        False, "--exact_match", help="Source Redis stream exact matching (default: False)"
    ),
    fuzzy_match: str | None = typer.Option(None, "--fuzzy_match", help="Source Redis stream fuzzy matching pattern"),
    overwrite: bool = typer.Option(False, "--overwrite", help="Whether to overwrite target streams (default: False)"),
    dry_run: bool = typer.Option(False, "--dry_run", help="Show operations without actually executing them"),
):
    """
    Batch restore Redis Stream data.

    Supports two matching modes:
    1. Exact match: Query fault_injection_schedules table from database to get corresponding trace_id
    2. Fuzzy match: Use Redis KEYS command to find streams matching specified pattern
    """
    # Parameter validation: must provide one, and cannot provide both
    if not exact_match and not fuzzy_match:
        console.print("[red]Error: Must select one matching mode[/red]")
        console.print(
            "[yellow]Use --exact_match for database exact matching, or --fuzzy_match for fuzzy matching[/yellow]"
        )
        raise typer.Exit(code=1)

    if exact_match and fuzzy_match:
        console.print("[red]Error: Cannot use both matching modes simultaneously[/red]")
        console.print("[yellow]Please choose either --exact_match or --fuzzy_match[/yellow]")
        raise typer.Exit(code=1)

    client = Client(source_url, target_url)
    client.copy_streams(exact_match, fuzzy_match, overwrite, dry_run)


if __name__ == "__main__":
    app()
