import psycopg
from psycopg.rows import dict_row
from rcabench.rcabench import RCABenchSDK
from rcabench.openapi.api_client import ApiClient, Configuration
from rcabench.openapi import (
    DatasetApi,
    AlgorithmApi,
    DtoExecutionPayload,
    DtoAlgorithmItem,
)
from rcabench.openapi.models.dto_dataset_build_payload import DtoDatasetBuildPayload
import time
import typer
import os
import shutil
import json
import uuid
from datetime import datetime

app = typer.Typer()


def connect_postgresql(host: str, user: str, password: str, dbname: str, port: int):
    return psycopg.connect(
        host=host,
        user=user,
        password=password,
        dbname=dbname,
        port=port,
    )


@app.command()
def dataset(
    base_url: str = typer.Option(
        "http://10.10.10.220:32080", help="RCABench 服务的 base URL"
    ),
    db_host: str = typer.Option("10.10.10.220", help="PostgreSQL 数据库主机"),
    db_user: str = typer.Option("postgres", help="PostgreSQL 用户名"),
    db_password: str = typer.Option("yourpassword", help="PostgreSQL 密码"),
    db_name: str = typer.Option("rcabench", help="PostgreSQL 数据库名"),
    db_port: int = typer.Option(32432, help="PostgreSQL 端口"),
    sleep_time: int = typer.Option(30, help="每次提交后的等待时间（秒）"),
):
    configuration: Configuration = Configuration(host=base_url)

    with ApiClient(configuration=configuration) as client:
        api = DatasetApi(api_client=client)
        try:
            with connect_postgresql(
                db_host, db_user, db_password, db_name, db_port
            ) as connection:
                print("✅ 成功连接到 PostgreSQL")

                with connection.cursor(row_factory=dict_row) as cursor:
                    # 获取版本信息
                    cursor.execute("SELECT VERSION() as version")
                    version_info = cursor.fetchone()
                    assert version_info, "未能获取 PostgreSQL版本信息"
                    print(f"📋 PostgreSQL版本: {version_info['version']}")

                    # 执行主查询
                    query = """
                    SELECT id, injection_name
                    FROM fault_injection_schedules
                    WHERE status = 3
                    ORDER BY id DESC
                    """

                    cursor.execute(query)
                    rows = cursor.fetchall()

                print(f"📋 查询结果：找到 {len(rows)} 条记录")

                for index, row in enumerate(rows, 1):
                    injection_id = row["id"]
                    injection_name = row["injection_name"]

                    print(
                        f"处理第 {index}/{len(rows)} 条：ID={injection_id}, Name={injection_name}"
                    )

                    try:
                        namespace = injection_name.split("-")[0]
                        print(f"  提取的命名空间: {namespace}")

                        resp = api.api_v1_datasets_post(
                            body=[
                                DtoDatasetBuildPayload(
                                    benchmark="clickhouse",
                                    name=injection_name,
                                    pre_duration=4,
                                    env_vars={
                                        "NAMESPACE": namespace,
                                    },
                                )
                            ]
                        )

                        print(f"  🔄 提交数据集成功：{resp}")

                    except Exception as submit_error:
                        print(f"  ❌ 提交数据集失败: {submit_error}")
                        continue

                    print(f"  ⏳ 等待 {sleep_time} 秒...")
                    time.sleep(sleep_time)

        except psycopg.Error as e:
            print(f"❌ PostgreSQL错误：{e}")
            raise typer.Exit(1)

        except Exception as e:
            print(f"❌ 其他错误：{e}")
            raise typer.Exit(1)


@app.command()
def detector(
    base_url: str = typer.Option(
        "http://10.10.10.220:32080", help="RCABench 服务的 base URL"
    ),
    db_host: str = typer.Option("10.10.10.220", help="PostgreSQL 数据库主机"),
    db_user: str = typer.Option("postgres", help="PostgreSQL 用户名"),
    db_password: str = typer.Option("yourpassword", help="PostgreSQL 密码"),
    db_name: str = typer.Option("rcabench", help="PostgreSQL 数据库名"),
    db_port: int = typer.Option(32432, help="PostgreSQL 端口"),
    sleep_time: int = typer.Option(10, help="每次提交后的等待时间（秒）"),
    detector_image: str = typer.Option("detector", help="检测器镜像名称"),
    # detector_tag: str = typer.Option("latest", help="检测器镜像标签"),
):
    configuration: Configuration = Configuration(host=base_url)

    with ApiClient(configuration=configuration) as client:
        api = AlgorithmApi(api_client=client)

        try:
            with connect_postgresql(
                db_host, db_user, db_password, db_name, db_port
            ) as connection:
                print("✅ 成功连接到 PostgreSQL")

                with connection.cursor(row_factory=dict_row) as cursor:
                    # 获取版本信息
                    cursor.execute("SELECT VERSION() as version")
                    version_info = cursor.fetchone()
                    assert version_info, "未能获取 PostgreSQL版本信息"
                    print(f"📋 PostgreSQL版本: {version_info['version']}")

                    query = """
                    SELECT id, injection_name 
                    FROM fault_injection_schedules
                    WHERE id NOT IN (
                        SELECT DISTINCT fis.id
                        FROM fault_injection_schedules fis 
                        JOIN execution_results er ON fis.injection_name = er.dataset
                        JOIN detectors d ON er.id = d.execution_id
                    ) AND status = 4
                    ORDER BY id DESC
                    """

                    cursor.execute(query)
                    rows = cursor.fetchall()

                print(f"📋 查询结果：找到 {len(rows)} 条记录")

                for index, row in enumerate(rows, 1):
                    injection_id = row["id"]
                    injection_name = row["injection_name"]

                    print(
                        f"处理第 {index}/{len(rows)} 条：ID={injection_id}, Name={injection_name}"
                    )

                    try:
                        resp = api.api_v1_algorithms_post(
                            body=[
                                DtoExecutionPayload(
                                    algorithm=DtoAlgorithmItem(name=detector_image),
                                    dataset=injection_name,
                                )
                            ]
                        )
                        print(f"  🔄 提交检测器成功：{resp}")

                    except Exception as submit_error:
                        print(f"  ❌ 提交检测器失败: {submit_error}")
                        continue

                    print(f"  ⏳ 等待 {sleep_time} 秒...")
                    time.sleep(sleep_time)

        except psycopg.Error as e:
            print(f"❌ PostgreSQL错误：{e}")
            raise typer.Exit(1)

        except Exception as e:
            print(f"❌ 其他错误：{e}")
            raise typer.Exit(1)


@app.command()
def align_db(
    db_host: str = typer.Option("10.10.10.220", help="PostgreSQL 数据库主机"),
    db_user: str = typer.Option("postgres", help="PostgreSQL 用户名"),
    db_password: str = typer.Option("yourpassword", help="PostgreSQL 密码"),
    db_name: str = typer.Option("rcabench", help="PostgreSQL 数据库名"),
    db_port: int = typer.Option(32432, help="PostgreSQL 端口"),
):
    path = "/mnt/jfs/rcabench_dataset"

    # 获取本地目录列表
    local_datasets = []
    if os.path.exists(path):
        local_datasets = [
            entry
            for entry in os.listdir(path)
            if os.path.isdir(os.path.join(path, entry))
        ]
        print(f"📁 本地找到 {len(local_datasets)} 个数据集目录")
    else:
        print(f"⚠️ 路径不存在: {path}")
        return

    with connect_postgresql(
        db_host, db_user, db_password, db_name, db_port
    ) as connection:
        with connection.cursor(row_factory=dict_row) as cursor:
            cursor.execute("SELECT VERSION() as version")
            version_info = cursor.fetchone()
            assert version_info, "未能获取 PostgreSQL版本信息"
            print(f"📋 PostgreSQL版本: {version_info['version']}")

            query = """
            SELECT id, injection_name 
            FROM fault_injection_schedules
            ORDER BY id DESC
            """
            cursor.execute(query)
            rows = cursor.fetchall()

            print(f"📋 数据库查询结果：找到 {len(rows)} 条记录")

            # 检查数据库中的记录是否在本地存在，如果不存在则删除
            deleted_count = 0
            database_datasets = []
            for row in rows:
                injection_id = row["id"]
                injection_name = row["injection_name"]
                database_datasets.append(injection_name)

                if injection_name not in local_datasets:
                    try:
                        delete_query = """
                        DELETE FROM fault_injection_schedules 
                        WHERE id = %s
                        """
                        cursor.execute(delete_query, (injection_id,))
                        print(
                            f"🗑️ 删除数据库记录: ID={injection_id}, Name={injection_name}"
                        )
                        deleted_count += 1
                    except Exception as e:
                        print(f"❌ 删除记录失败 ID={injection_id}: {e}")

            connection.commit()
            print(f"✅ 总共删除了 {deleted_count} 条数据库记录")

            # 检查本地数据集是否在数据库中存在，如果不存在则从injection.json添加记录
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

                            # 生成新的task_id
                            new_task_id = str(uuid.uuid4())

                            # 构建插入语句
                            insert_query = """
                            INSERT INTO fault_injection_schedules (
                                task_id, fault_type, display_config, engine_config, 
                                pre_duration, start_time, end_time, status, 
                                description, benchmark, injection_name, labels,
                                created_at, updated_at
                            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                            """

                            # 准备数据
                            values = (
                                new_task_id,
                                injection_data.get("fault_type"),
                                injection_data.get("display_config"),
                                injection_data.get("engine_config"),
                                injection_data.get("pre_duration"),
                                injection_data.get("start_time"),
                                injection_data.get("end_time"),
                                injection_data.get("status"),
                                injection_data.get("description"),
                                injection_data.get("benchmark"),
                                injection_data.get("injection_name"),
                                json.dumps(injection_data.get("labels", {})),
                                injection_data.get("created_at"),
                                injection_data.get("updated_at"),
                            )

                            cursor.execute(insert_query, values)
                            print(
                                f"➕ 添加数据库记录: Name={local_dataset}, TaskID={new_task_id}"
                            )
                            added_count += 1

                        except Exception as e:
                            print(f"❌ 添加记录失败 {local_dataset}: {e}")
                    else:
                        print(f"⚠️ 缺少injection.json文件: {injection_json_path}")

            connection.commit()
            print(f"✅ 总共添加了 {added_count} 条数据库记录")


if __name__ == "__main__":
    app()
