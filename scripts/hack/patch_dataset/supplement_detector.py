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
            cursor.execute("""SELECT id, injection_name FROM fault_injection_schedules
WHERE id not in (
    SELECT DISTINCT fis.id
    FROM ((`rcabench`.`fault_injection_schedules` `fis` join `rcabench`
        .`execution_results` `er`
           ON (`fis`.`id` = `er`.`dataset`)) JOIN `rcabench`.`detectors` `d`
          ON (`er`.`id` = `d`.`execution_id`))
) AND status=4
ORDER BY id DESC;
""")

            rows = cursor.fetchall()

            print("📋 查询结果：")
            for row in rows:
                print(row[1])
                resp = sdk.algorithm.submit(
                    [
                        {
                            "image": "detector",
                            "dataset": row[1],
                            "tag": "latest",
                            "env_vars": {},
                        }
                    ]
                )
                print(f"🔄 提交数据集：{resp}")
                time.sleep(10)

    except Error as e:
        print(f"❌ 查询失败：{e}")

    finally:
        if connection.is_connected():
            cursor.close()
            connection.close()
            print("🔌 已关闭连接")


if __name__ == "__main__":
    query_mariadb()
