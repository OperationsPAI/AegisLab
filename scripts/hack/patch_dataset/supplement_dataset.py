import pymysql
from rcabench.rcabench import RCABenchSDK
import time

sdk = RCABenchSDK(base_url="http://127.0.0.1:8082")


def query_mariadb():
    connection = None
    cursor = None
    
    try:
        connection = pymysql.connect(
            host='127.0.0.1',
            user='root',
            password='yourpassword',
            database='rcabench',
            port=3306,
            charset='utf8mb4',
            autocommit=True,
            cursorclass=pymysql.cursors.DictCursor  # 使用字典游标，更方便
        )
        
        print("✅ 成功连接到 MariaDB")
        
        with connection.cursor() as cursor:
            # 获取版本信息
            cursor.execute("SELECT VERSION() as version")
            version_info = cursor.fetchone()
            print(f"📋 MariaDB版本: {version_info['version']}")
            
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
                injection_id = row['id']
                injection_name = row['injection_name']
                
                print(f"处理第 {index}/{len(rows)} 条：ID={injection_id}, Name={injection_name}")
                
                try:
                    namespace = injection_name.split("-")[0]
                    print(f"  提取的命名空间: {namespace}")
                    
                    resp = sdk.dataset.submit([
                        {
                            "benchmark": "clickhouse",
                            "name": injection_name,
                            "pre_duration": 4,
                            "env_vars": {
                                "NAMESPACE": namespace,
                            },
                        }
                    ])
                    print(f"  🔄 提交数据集成功：{resp}")
                    
                except Exception as submit_error:
                    print(f"  ❌ 提交数据集失败: {submit_error}")
                    continue
                
                print(f"  ⏳ 等待 20 秒...")
                time.sleep(20)
                
    except pymysql.Error as e:
        print(f"❌ MariaDB错误：{e}")
        
    except Exception as e:
        print(f"❌ 其他错误：{e}")
        
    finally:
        if connection:
            connection.close()
            print("🔌 已关闭MariaDB连接")


if __name__ == "__main__":
    query_mariadb()