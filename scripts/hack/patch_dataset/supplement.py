#!/usr/bin/env -S uv run -s
import mysql.connector
from mysql.connector import Error
from rcabench.rcabench import RCABenchSDK
from rcabench.openapi.api_client import ApiClient, Configuration
from rcabench.openapi import (
    DatasetApi,
    AlgorithmApi,
    DtoExecutionPayload,
    DtoAlgorithmItem,
    DtoSubmitDatasetBuildingReq,
    DtoSubmitExecutionReq,
)
from rcabench.openapi.models.dto_dataset_build_payload import DtoDatasetBuildPayload
import time
import typer
import os
import json
from datetime import datetime
from typing import Dict, Any, Optional

app = typer.Typer()


def connect_mysql(host: str, user: str, password: str, dbname: str, port: int):
    return mysql.connector.connect(
        host=host,
        user=user,
        password=password,
        database=dbname,
        port=port,
    )


@app.command()
def dataset(
    base_url: str = typer.Option(
        "http://10.10.10.220:32080", help="Base URL of RCABench service"
    ),
    db_host: str = typer.Option("10.10.10.220", help="MySQL database host"),
    db_user: str = typer.Option("root", help="MySQL username"),
    db_password: str = typer.Option("yourpassword", help="MySQL password"),
    db_name: str = typer.Option("rcabench", help="MySQL database name"),
    db_port: int = typer.Option(32206, help="MySQL port"),
    sleep_time: int = typer.Option(
        30, help="Wait time after each submission (seconds)"
    ),
):
    configuration: Configuration = Configuration(host=base_url)

    with ApiClient(configuration=configuration) as client:
        api = DatasetApi(api_client=client)
        try:
            with connect_mysql(
                db_host, db_user, db_password, db_name, db_port
            ) as connection:
                print("✅ Successfully connected to MySQL")

                with connection.cursor(dictionary=True) as cursor:
                    # Get version information
                    cursor.execute("SELECT VERSION() as version")
                    version_info: Optional[Dict[str, Any]] = cursor.fetchone()  # type: ignore
                    assert version_info, "Failed to get MySQL version information"
                    print(f"📋 MySQL version: {version_info['version']}")

                    # Execute main query
                    query = """
                    SELECT id, injection_name
                    FROM fault_injection_schedules
                    WHERE status = 3
                    ORDER BY id DESC
                    """

                    cursor.execute(query)
                    rows = cursor.fetchall()

                print(f"📋 Query result: found {len(rows)} records")

                for index, row in enumerate(rows, 1):
                    injection_id = row["id"]  # type: ignore
                    injection_name = str(row["injection_name"])  # type: ignore

                    print(
                        f"Processing {index}/{len(rows)}: ID={injection_id}, Name={injection_name}"
                    )

                    try:
                        namespace = injection_name.split("-")[0]
                        print(f"  Extracted namespace: {namespace}")

                        resp = api.api_v1_datasets_post(
                            body=DtoSubmitDatasetBuildingReq(
                                project_name="pair_diagnosis",
                                payloads=[
                                    DtoDatasetBuildPayload(
                                        benchmark="clickhouse",
                                        name=injection_name,
                                        pre_duration=4,
                                        env_vars={
                                            "NAMESPACE": namespace,
                                        },
                                    ),
                                ],
                            ),
                        )

                        print(f"  🔄 Dataset submission successful: {resp}")

                    except Exception as submit_error:
                        print(f"  ❌ Dataset submission failed: {submit_error}")
                        continue

                    print(f"  ⏳ Waiting {sleep_time} seconds...")
                    time.sleep(sleep_time)

        except Error as e:
            print(f"❌ MySQL error: {e}")
            raise typer.Exit(1)

        except Exception as e:
            print(f"❌ Other error: {e}")
            raise typer.Exit(1)


@app.command()
def detector(
    base_url: str = typer.Option(
        "http://10.10.10.220:32080", help="Base URL of RCABench service"
    ),
    db_host: str = typer.Option("10.10.10.220", help="MySQL database host"),
    db_user: str = typer.Option("root", help="MySQL username"),
    db_password: str = typer.Option("yourpassword", help="MySQL password"),
    db_name: str = typer.Option("rcabench", help="MySQL database name"),
    db_port: int = typer.Option(32206, help="MySQL port"),
    sleep_time: int = typer.Option(
        10, help="Wait time after each submission (seconds)"
    ),
    detector_image: str = typer.Option("detector", help="Detector image name"),
    # detector_tag: str = typer.Option("latest", help="Detector image tag"),
):
    configuration: Configuration = Configuration(host=base_url)

    with ApiClient(configuration=configuration) as client:
        api = AlgorithmApi(api_client=client)

        try:
            with connect_mysql(
                db_host, db_user, db_password, db_name, db_port
            ) as connection:
                print("✅ Successfully connected to MySQL")

                with connection.cursor(dictionary=True) as cursor:
                    # Get version information
                    cursor.execute("SELECT VERSION() as version")
                    version_info: Optional[Dict[str, Any]] = cursor.fetchone()  # type: ignore
                    assert version_info, "Failed to get MySQL version information"
                    print(f"📋 MySQL version: {version_info['version']}")

                    query = """
                    SELECT id, injection_name 
                    FROM fault_injection_schedules
                    WHERE id NOT IN (
                        SELECT DISTINCT fis.id
                        FROM fault_injection_schedules fis 
                        JOIN execution_results er ON fis.id = er.datapack_id
                        JOIN detectors d ON er.id = d.execution_id
                    ) AND status = 4
                    ORDER BY id DESC
                    """

                    cursor.execute(query)
                    rows = cursor.fetchall()

                print(f"📋 Query result: found {len(rows)} records")

                for index, row in enumerate(rows, 1):
                    injection_id = row["id"]  # type: ignore
                    injection_name = str(row["injection_name"])  # type: ignore

                    print(
                        f"Processing {index}/{len(rows)}: ID={injection_id}, Name={injection_name}"
                    )

                    try:
                        resp = api.api_v1_algorithms_post(
                            body=DtoSubmitExecutionReq(
                                project_name="pair_diagnosis",
                                payloads=[
                                    DtoExecutionPayload(
                                        algorithm=DtoAlgorithmItem(name=detector_image),
                                        dataset=injection_name,
                                    )
                                ],
                            ),
                        )
                        print(f"  🔄 Detector submission successful: {resp}")

                    except Exception as submit_error:
                        print(f"  ❌ Detector submission failed: {submit_error}")
                        continue

                    print(f"  ⏳ Waiting {sleep_time} seconds...")
                    time.sleep(sleep_time)

        except Error as e:
            print(f"❌ MySQL error: {e}")
            raise typer.Exit(1)

        except Exception as e:
            print(f"❌ Other error: {e}")
            raise typer.Exit(1)


@app.command()
def detector_single(
    injection_name: str,
    base_url: str = typer.Option(
        "http://10.10.10.220:32080", help="Base URL of RCABench service"
    ),
    detector_image: str = typer.Option("detector", help="Detector image name"),
):
    configuration: Configuration = Configuration(host=base_url)

    with ApiClient(configuration=configuration) as client:
        api = AlgorithmApi(api_client=client)
        resp = api.api_v1_algorithms_post(
            body=DtoSubmitExecutionReq(
                project_name="pair_diagnosis",
                payloads=[
                    DtoExecutionPayload(
                        algorithm=DtoAlgorithmItem(name=detector_image),
                        dataset=injection_name,
                    )
                ],
            ),
        )


@app.command()
def align_db(
    db_host: str = typer.Option("10.10.10.220", help="MySQL database host"),
    db_user: str = typer.Option("root", help="MySQL username"),
    db_password: str = typer.Option("yourpassword", help="MySQL password"),
    db_name: str = typer.Option("rcabench", help="MySQL database name"),
    db_port: int = typer.Option(32206, help="MySQL port"),
):
    path = "/mnt/jfs/rcabench_dataset"

    # Get local directory list
    local_datasets = []
    if os.path.exists(path):
        local_datasets = [
            entry
            for entry in os.listdir(path)
            if os.path.isdir(os.path.join(path, entry))
        ]
        print(f"📁 Found {len(local_datasets)} local dataset directories")
    else:
        print(f"⚠️ Path does not exist: {path}")
        return

    with connect_mysql(db_host, db_user, db_password, db_name, db_port) as connection:
        with connection.cursor(dictionary=True) as cursor:
            cursor.execute("SELECT VERSION() as version")
            version_info: Optional[Dict[str, Any]] = cursor.fetchone()
            assert version_info, "Failed to get MySQL version information"
            print(f"📋 MySQL version: {version_info['version']}")

            query = """
            SELECT id, injection_name 
            FROM fault_injection_schedules
            ORDER BY id DESC
            """
            cursor.execute(query)
            rows = cursor.fetchall()

            print(f"📋 Database query result: found {len(rows)} records")

            # Check if database records exist locally, delete if not found
            deleted_count = 0
            database_datasets = []
            for row in rows:
                injection_id = row["id"]
                injection_name = str(row["injection_name"])
                database_datasets.append(injection_name)

                if injection_name not in local_datasets:
                    try:
                        # Delete dependent table data (in foreign key dependency order)

                        # 1. Delete detectors table
                        cursor.execute(
                            """DELETE d FROM detectors d 
JOIN execution_results er ON d.execution_id = er.id 
WHERE er.datapack_id = %s
                        """,
                            (injection_id,),
                        )

                        # 2. Delete execution_result_labels table
                        cursor.execute(
                            """DELETE erl FROM execution_result_labels erl
JOIN execution_results er ON erl.execution_id = er.id 
WHERE er.datapack_id = %s
                        """,
                            (injection_id,),
                        )

                        # 3. Delete granularity_results table
                        cursor.execute(
                            """DELETE gr FROM granularity_results gr
JOIN execution_results er ON gr.execution_id = er.id 
WHERE er.datapack_id = %s
                        """,
                            (injection_id,),
                        )

                        # 4. Delete execution_results table
                        cursor.execute(
                            """
DELETE FROM execution_results WHERE datapack_id = %s
                        """,
                            (injection_id,),
                        )

                        # 5. Delete fault_injection_labels table
                        cursor.execute(
                            """
DELETE FROM fault_injection_labels WHERE fault_injection_id = %s
                        """,
                            (injection_id,),
                        )

                        # 6. Delete dataset_fault_injections table
                        cursor.execute(
                            """
DELETE FROM dataset_fault_injections WHERE fault_injection_id = %s
                        """,
                            (injection_id,),
                        )

                        # 7. Finally delete main table fault_injection_schedules
                        cursor.execute(
                            """
DELETE FROM fault_injection_schedules WHERE id = %s
                        """,
                            (injection_id,),
                        )

                        # Commit transaction
                        connection.commit()
                        print(
                            f"🗑️ Deleted database record: ID={injection_id}, Name={injection_name}"
                        )
                        deleted_count += 1
                    except Exception as e:
                        # Rollback transaction
                        connection.rollback()
                        print(f"❌ Failed to delete record ID={injection_id}: {e}")

            print(f"✅ Total deleted {deleted_count} database records")

            # Check if local datasets exist in database, add from injection.json if not found
            added_count = 0
            for local_dataset in local_datasets:
                if local_dataset not in database_datasets:
                    injection_json_path = os.path.join(
                        path, local_dataset, "injection.json"
                    )
                    if os.path.exists(injection_json_path):
                        try:
                            with open(injection_json_path, "r", encoding="utf-8") as f:
                                injection_data = json.load(f)

                            # Generate new task_id - use NULL instead of UUID due to foreign key constraints
                            new_task_id = None

                            # Build insert statement
                            insert_query = """
                            INSERT INTO fault_injection_schedules (
                                task_id, fault_type, display_config, engine_config, 
                                pre_duration, start_time, end_time, status, 
                                description, benchmark, injection_name,
                                created_at, updated_at
                            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                            """

                            # Prepare data and perform type conversion
                            def safe_get(
                                data: Dict[str, Any], key: str, default: Any = None
                            ) -> Any:
                                value = data.get(key, default)
                                if value is None:
                                    return None
                                return value

                            def parse_timestamp(timestamp_str: Any) -> Any:
                                if timestamp_str is None:
                                    return None
                                try:
                                    # Try to parse timestamp string
                                    if isinstance(timestamp_str, str):
                                        return datetime.fromisoformat(
                                            timestamp_str.replace("Z", "+00:00")
                                        )
                                    return timestamp_str
                                except Exception:
                                    return None

                            values = (
                                new_task_id,  # Set task_id to NULL
                                safe_get(injection_data, "fault_type"),
                                safe_get(injection_data, "display_config"),
                                safe_get(injection_data, "engine_config"),
                                safe_get(injection_data, "pre_duration"),
                                parse_timestamp(safe_get(injection_data, "start_time")),
                                parse_timestamp(safe_get(injection_data, "end_time")),
                                4,
                                safe_get(injection_data, "description"),
                                safe_get(injection_data, "benchmark"),
                                safe_get(injection_data, "injection_name"),
                                parse_timestamp(safe_get(injection_data, "created_at")),
                                parse_timestamp(safe_get(injection_data, "updated_at")),
                            )

                            cursor.execute(insert_query, values)
                            print(f"➕ Added database record: Name={local_dataset}")
                            added_count += 1

                        except Exception as e:
                            print(f"❌ Failed to add record {local_dataset}: {e}")
                            # Rollback current transaction to avoid affecting subsequent operations
                            connection.rollback()
                            # Restart transaction
                            connection.commit()
                    else:
                        print(f"⚠️ Missing injection.json file: {injection_json_path}")

            connection.commit()
            print(f"✅ Total added {added_count} database records")


if __name__ == "__main__":
    app()
