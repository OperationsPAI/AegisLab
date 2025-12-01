import redis
import sqlalchemy
from redis import Redis
from sqlalchemy import create_engine, text
from sqlalchemy.orm import Session, sessionmaker

from src.backup.mysql import local_mysql_config, remote_mysql_config
from src.common.common import SourceType, console, settings

HASH_PATTERN = "injection:algorithms"
STREAM_PATTERN = "trace:*:log"
TRACE_LOG_KEY = "trace:{}:log"

local_redis_url = f"redis://{settings.redis.host}:{settings.redis.port}"
settings.setenv("remote")
remote_redis_url = f"redis://{settings.redis.host}:{settings.redis.port}"


class RedisClient:
    def __init__(self, src: SourceType):
        if src == SourceType.LOCAL:
            self.dst = SourceType.REMOTE
            self.src_mysql_config = local_mysql_config
            self.dst_mysql_config = remote_mysql_config
            src_redis_url = local_redis_url
            dst_redis_url = remote_redis_url
        elif src == SourceType.REMOTE:
            self.dst = SourceType.LOCAL
            self.src_mysql_config = remote_mysql_config
            self.dst_mysql_config = local_mysql_config
            src_redis_url = remote_redis_url
            dst_redis_url = local_redis_url

        self.source_redis, self.target_redis = self._connect_redis(
            src_redis_url, dst_redis_url
        )
        self.db_session = self._connect_database()

    def _connect_redis(self, source_url: str, target_url: str) -> tuple[Redis, Redis]:
        """
        Connect to source and target Redis instances.
        """
        console.print(f"[cyan]Connecting to source Redis: {source_url}[/cyan]")
        console.print(f"[cyan]Connecting to target Redis: {target_url}[/cyan]")

        try:
            source_redis: Redis = redis.from_url(source_url, decode_responses=True)
            target_redis: Redis = redis.from_url(target_url, decode_responses=True)

            console.print("[cyan]Testing source Redis connection...[/cyan]")
            source_redis.ping()
            console.print(
                "[bold green]✅ Source Redis connection successful[/bold green]"
            )

            console.print("[cyan]Testing target Redis connection...[/cyan]")
            target_redis.ping()
            console.print(
                "[bold green]✅ Target Redis connection successful[/bold green]"
            )

            return source_redis, target_redis

        except redis.ConnectionError as e:
            console.print(f"[bold red]Redis connection failed: {e}[/bold red]")
            raise SystemExit(1)

    def _connect_database(self) -> Session:
        """
        Connect to the MySQL database.
        """
        try:
            db_url = f"mysql+pymysql://{self.src_mysql_config.user}:{self.src_mysql_config.password}@{self.src_mysql_config.host}:{self.src_mysql_config.port}/{self.src_mysql_config.db}"
            engine = create_engine(db_url, echo=False, pool_pre_ping=True)

            Session = sessionmaker(bind=engine)
            session = Session()

            # Test connection
            result = session.execute(text("SELECT version()"))
            version = result.scalar()
            console.print("[bold green]✅ Database connection successful[/bold green]")
            console.print(f"[dim]Version: {version[:50]}...[/dim]")  # type: ignore

            return session

        except sqlalchemy.except_all.SQLAlchemyError as e:
            console.print(f"[bold red]❌ Database connection failed: {e}[/bold red]")
            raise SystemExit(1)

    def _read_hashes_fuzzy(self, pattern: str) -> list[str]:
        """
        Find hash tables matching the given pattern using fuzzy matching.
        """
        console.print(
            f"[bold blue]Searching for hash tables matching '{pattern}'...[/bold blue]"
        )

        matching_keys = self.source_redis.keys(pattern)
        if not matching_keys:
            console.print(f"[bold red]No keys found matching '{pattern}'[/bold red]")
            raise SystemExit(1)

        hashes = []
        for key in matching_keys:  # type: ignore
            if self.source_redis.type(key) == "hash":
                hashes.append(key)

        if not hashes:
            console.print(
                "[bold yellow]No hash table data found among matching keys[/bold yellow]"
            )
            raise SystemExit(1)

        console.print(
            f"[bold green]✅ Found {len(hashes)} matching hash tables[/bold green]"
        )
        return hashes

    def _read_streams_exact(self) -> list[str]:
        """
        Get stream names by querying the database for exact matches.
        """
        console.print("[bold blue]Executing query to get stream names...[/bold blue]")

        query = """
        SELECT DISTINCT t.trace_id
        FROM fault_injection fis
        INNER JOIN tasks t ON fis.task_id = t.id
        WHERE fis.task_id IS NOT NULL;
        """

        try:
            console.print(f"[dim]SQL: {query}[/dim]")

            result = self.db_session.execute(text(query))
            results = result.fetchall()
            if not results:
                console.print("[bold yellow]Query returned no results[/bold yellow]")
                return []

            streams = []
            for row in results:
                trace_id = row[0]  # SQLAlchemy result is in tuple format
                if trace_id:
                    streams.append(TRACE_LOG_KEY.format(trace_id))

            console.print(
                f"[bold green]✅ Retrieved {len(streams)} stream names from database[/bold green]"
            )
            return streams

        except sqlalchemy.except_all.SQLAlchemyError as e:
            console.print(f"[red]❌ Database query failed: {e}[/red]")
            raise SystemExit(1)

    def _read_streams_fuzzy(self, pattern: str) -> list[str]:
        """Find streams matching the given pattern using fuzzy matching.

        Args:
            pattern: Redis key pattern to match

        Returns:
            List of stream keys that match the pattern

        Raises:
            typer.Exit: If no matching streams are found
        """
        console.print(
            f"[bold blue]Searching for streams matching '{pattern}'...[/bold blue]"
        )

        matching_keys = self.source_redis.keys(pattern)
        if not matching_keys:
            console.print(f"[bold red]No streams found matching '{pattern}'[/bold red]")
            raise SystemExit(1)

        # Filter for actual stream type
        streams = []
        for key in matching_keys:  # type: ignore
            if self.source_redis.type(key) == "stream":
                streams.append(key)

        if not streams:
            console.print(
                "[bold yellow]No stream type data found among matching keys[/bold yellow]"
            )
            raise SystemExit(1)

        console.print(
            f"[bold green]✅ Found {len(streams)} matching streams[/bold green]"
        )
        return streams

    def copy_hashes(self, force: bool = False, dry_run: bool = False) -> None:
        """
        Copy hash tables from source to target Redis instance.
        """
        hashes = self._read_hashes_fuzzy(HASH_PATTERN)

        if dry_run:
            console.print(
                "[bold yellow]Dry run mode, no data will be actually copied[/bold yellow]"
            )
            return

        console.print("[bold blueStarting batch copy...[/bold blue]")

        success_count = 0
        failed_count = 0
        skipped_count = 0

        for i, source_key in enumerate(hashes, 1):
            try:
                target_key = source_key

                # Check source hash table
                if not self.source_redis.exists(source_key):
                    console.print(
                        f"[bold yellow][{i}/{len(hashes)}] Skipping non-existent: {source_key}[/bold yellow]"
                    )
                    skipped_count += 1
                    continue

                hash_length = self.source_redis.hlen(source_key)
                if hash_length == 0:
                    console.print(
                        f"[bold yellow][{i}/{len(hashes)}] Skipping empty hash table: {source_key}[/bold yellow]"
                    )
                    skipped_count += 1
                    continue

                # Check target hash table
                target_exists = self.target_redis.exists(target_key)
                if target_exists and not force:
                    target_length = self.target_redis.hlen(target_key)
                    console.print(
                        f"[bold yellow][{i}/{len(hashes)}] Skipping existing: {source_key} ({target_length} fields)[/bold yellow]"
                    )
                    skipped_count += 1
                    continue

                # Display current hash table being processed
                console.print(
                    f"[bold blue][{i}/{len(hashes)}] Copying: {source_key} ({hash_length} fields)[/bold blue]"
                )

                # Delete target hash table if force is needed
                if target_exists and force:
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
                        f"[dim]  [bold green]✓[/bold green] {len(all_fields)} fields[dim]"  # type: ignore
                    )
                except Exception as e:
                    failed_count += 1
                    console.print(
                        f"[dim] [bold red]  ✗ Copy failed: {e}[/bold red][dim]"
                    )

                # Show progress every 10 hash tables
                if i % 10 == 0:
                    console.print(
                        f"[dim]Progress: {i}/{len(hashes)} ({success_count} success, {failed_count} failed, {skipped_count} skipped)[/dim]"  # noqa: E501
                    )

            except Exception as e:
                failed_count += 1
                console.print(
                    f"[bold red][{i}/{len(hashes)}] Processing failed: {source_key} - {e}[/bold red]"
                )

        # Display final results
        console.print()
        console.print("[bold green]Batch copy completed[/bold green]")
        console.print(
            f"[bold green]✅ Success: {success_count} hash tables[/bold green]"
        )

        if failed_count > 0:
            console.print(f"[bold red]❌ Failed: {failed_count} hash tables[/bold red]")

        if skipped_count > 0:
            console.print(
                f"[bold yellow]🚫 Skipped: {skipped_count} hash tables[/bold yellow]"
            )

        console.print(
            f"[bold green]Total processed: {len(hashes)} hash tables[/bold green]"
        )

    def copy_streams(
        self,
        exact_match: bool = False,
        force: bool = False,
        dry_run: bool = False,
    ) -> None:
        """
        Copy Redis streams from source to target instance.
        """
        if exact_match:
            streams = self._read_streams_exact()
        else:
            streams = self._read_streams_fuzzy(STREAM_PATTERN)

        if dry_run:
            console.print(
                "[bold yellow]Dry run mode, no data will be actually copied[/bold yellow]"
            )
            return

        console.print("[bold blue]Starting batch copy...[/bold blue]")

        success_count = 0
        failed_count = 0
        skipped_count = 0

        for i, source_key in enumerate(streams, 1):
            try:
                target_key = source_key

                # Check source stream
                if not self.source_redis.exists(source_key):
                    console.print(
                        f"[bold yellow][{i}/{len(streams)}] Skipping non-existent: {source_key}[/bold yellow]"
                    )
                    skipped_count += 1
                    continue

                stream_length = self.source_redis.xlen(source_key)
                if stream_length == 0:
                    console.print(
                        f"[bold yellow][{i}/{len(streams)}] Skipping empty stream: {source_key}[/bold yellow]"
                    )
                    skipped_count += 1
                    continue

                # Check target stream
                target_exists = self.target_redis.exists(target_key)
                if target_exists and not force:
                    target_length = self.target_redis.xlen(target_key)
                    console.print(
                        f"[bold yellow][{i}/{len(streams)}] Skipping existing: {source_key} ({target_length} records)[/bold yellow]"  # noqa: E501
                    )
                    skipped_count += 1
                    continue

                # Display current stream being processed
                console.print(
                    f"[bold blue][{i}/{len(streams)}] Copying: {source_key} ({stream_length} records)[/bold blue]"
                )

                # Delete target stream if force is needed
                if target_exists and force:
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
                    status_msg = "[bold green]✅[/bold green]"
                    if message_failed_count > 0:
                        status_msg += f" [bold yellow]({message_failed_count} failed)[/bold yellow]"

                    console.print(f"[dim]  {status_msg} {copied_count} records[/dim]")
                else:
                    failed_count += 1
                    console.print("[bold red]❌ Copy failed[/bold red]")

                # Show progress every 10 streams
                if i % 10 == 0:
                    console.print(
                        f"[dim]Progress: {i}/{len(streams)} ({success_count} success, {failed_count} failed, {skipped_count} skipped)[/dim]"  # noqa: E501
                    )

            except Exception as e:
                failed_count += 1
                console.print(
                    f"[bold red][{i}/{len(streams)}] Processing failed: {source_key} - {e}[/bold red]"
                )

        # Display final results
        console.print()
        console.print("[bold green]Batch copy completed[/bold green]")
        console.print(f"[bold green]✅ Success: {success_count} streams[/bold green]")

        if failed_count > 0:
            console.print(f"[bold red]❌ Failed: {failed_count} streams[/bold red]")

        if skipped_count > 0:
            console.print(
                f"[bold yellow]🚫 Skipped: {skipped_count} streams[/bold yellow]"
            )

        console.print(
            f"[bold green]Total processed: {len(streams)} streams[/bold green]"
        )
