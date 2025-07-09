import psycopg
from psycopg.rows import dict_row
from rcabench.rcabench import RCABenchSDK
from rcabench.openapi.api_client import ApiClient, Configuration
from rcabench.openapi import DatasetApi, AlgorithmApi
from rcabench.openapi.models.dto_dataset_build_payload import DtoDatasetBuildPayload
from rcabench.openapi.models.dto_algorithm_execution_payload import (
    DtoAlgorithmExecutionPayload,
)
import time
import typer

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
    sleep_time: int = typer.Option(20, help="每次提交后的等待时间（秒）"),
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
                    WHERE created_at > '2025-06-14 00:00:00' 
                    AND status = 4
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
                                DtoAlgorithmExecutionPayload(
                                    algorithm=detector_image,
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

        print(f"📋 查询结果：找到 {len(rows)} 条记录")

    datasets = [row["injection_name"] for row in rows]

    path = "/mnt/jfs/rcabench_dataset"

    import os
    import shutil

    if os.path.exists(path):
        for entry in os.listdir(path):
            full_path = os.path.join(path, entry)
            if os.path.isdir(full_path) and entry not in datasets:
                print(f"🗑️ 删除多余目录: {full_path}")
                shutil.rmtree(full_path)
    else:
        print(f"⚠️ 路径不存在: {path}")


if __name__ == "__main__":
    app()
