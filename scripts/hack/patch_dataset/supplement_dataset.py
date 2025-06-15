import mysql.connector
from mysql.connector import Error
from rcabench.rcabench import RCABenchSDK
import time

sdk = RCABenchSDK(base_url="http://10.10.10.220:32080")


def query_mariadb():
    try:
        connection = mysql.connector.connect(
            host="10.10.10.220",
            user="root",
            password="yourpassword",
            database="rcabench",
            port=32336,
        )

        if connection.is_connected():
            print("✅ 成功连接到 MariaDB")

            cursor = connection.cursor()
            cursor.execute("""SELECT id, injection_name
FROM fault_injection_schedules s
WHERE  created_at > '2025-06-14 00:00:00' and (status=4)
ORDER BY id DESC;""")

            rows = cursor.fetchall()

            print("📋 查询结果：")
            for row in rows:
                print(row[1])
                resp = sdk.dataset.submit(
                    [
                        {
                            "benchmark": "clickhouse",
                            "name": row[1],
                            "pre_duration": 4,
                            "env_vars": {
                                "NAMESPACE": row[1].split("-")[0],
                            },
                        }
                    ]
                )
                print(f"🔄 提交数据集：{resp}")
                time.sleep(20)

    except Error as e:
        print(f"❌ 查询失败：{e}")

    finally:
        if connection.is_connected():
            cursor.close()
            connection.close()
            print("🔌 已关闭连接")


if __name__ == "__main__":
    query_mariadb()
