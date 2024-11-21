
import os
import pandas as pd
import matplotlib.pyplot as plt
from sklearn.preprocessing import MinMaxScaler

# 基本路径
addr = R'E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\cpu\checkoutservice-1011-1441'

# 文件夹路径
abnormal_folder = os.path.join(addr, 'abnormal', 'processed_metrics')
normal_folder = os.path.join(addr, 'normal', 'processed_metrics')
output_folder = os.path.join(addr, 'figure')

# 确保输出文件夹存在
os.makedirs(output_folder, exist_ok=True)

# 遍历 abnormal 文件夹中的所有 CSV 文件
for filename in os.listdir(abnormal_folder):
    if filename.endswith(".csv"):
        abnormal_file_path = os.path.join(abnormal_folder, filename)
        normal_file_path = os.path.join(normal_folder, filename)
        
        # 检查是否有对应的 normal 文件
        if not os.path.exists(normal_file_path):
            print(f"没有找到 {filename} 对应的 normal 文件，跳过此文件。")
            continue
        
        # 读取 abnormal 和 normal 文件
        abnormal_data = pd.read_csv(abnormal_file_path)
        normal_data = pd.read_csv(normal_file_path)
        
        # 确保 'TimeUnix' 列被解析为日期时间格式
        abnormal_data['TimeUnix'] = pd.to_datetime(abnormal_data['TimeUnix'])
        normal_data['TimeUnix'] = pd.to_datetime(normal_data['TimeUnix'])
        
        # 准备归一化处理
        scaler = MinMaxScaler()
        
        # 选择需要归一化的列
        metric_columns = [
            'k8s.pod.cpu.usage',
            'k8s.pod.cpu_limit_utilization',
            'k8s.pod.filesystem.usage',
            'k8s.pod.memory.usage',
            'k8s.pod.memory_limit_utilization',
            'k8s.pod.network.errors',
            'receive_bytes',
            'transmit_bytes',
            'k8s.pod.filesystem.utilization'
        ]
        
        # 检查两个数据集中是否包含所有所需的列
        available_columns_abnormal = [col for col in metric_columns if col in abnormal_data.columns]
        available_columns_normal = [col for col in metric_columns if col in normal_data.columns]
        
        # 保证两个数据集的度量列一致
        available_columns = list(set(available_columns_abnormal) & set(available_columns_normal))
        
        if len(available_columns) == 0:
            print(f"文件 {filename} 中没有找到任何共同的度量列，跳过此文件。")
            continue
        
        # 将 abnormal 和 normal 数据组合在一起用于归一化
        combined_data = pd.concat([abnormal_data[available_columns], normal_data[available_columns]], axis=0)
        
        # 对组合的数据进行归一化处理
        scaled_combined_data = scaler.fit_transform(combined_data)
        
        # 将归一化后的数据拆分为 abnormal 和 normal 数据
        abnormal_data[available_columns] = scaled_combined_data[:len(abnormal_data)]
        normal_data[available_columns] = scaled_combined_data[len(abnormal_data):]
        
        # 设置绘图
        plt.figure(figsize=(15, 10))
        
        # 定义每个度量的颜色
        colors = {
            'k8s.pod.cpu.usage': 'blue',
            'k8s.pod.cpu_limit_utilization': 'green',
            'k8s.pod.filesystem.usage': 'red',
            'k8s.pod.memory.usage': 'cyan',
            'k8s.pod.memory_limit_utilization': 'magenta',
            'k8s.pod.network.errors': 'yellow',
            'receive_bytes': 'black',
            'transmit_bytes': 'orange',
            'k8s.pod.filesystem.utilization': 'purple'
        }
        
        # 绘制 abnormal 和 normal 数据的折线图
        for metric in available_columns:
            color = colors.get(metric, 'gray')  # 如果没有定义颜色，使用灰色
            
            # 绘制 abnormal 数据
            plt.plot(abnormal_data['TimeUnix'].values, abnormal_data[metric].to_numpy(), 
                     label=f'abnormal_{metric}', color=color, linestyle='-', alpha=0.7)
            
            # 绘制 normal 数据
            plt.plot(normal_data['TimeUnix'].values, normal_data[metric].to_numpy(), 
                     label=f'normal_{metric}', color=color, linestyle='--', alpha=0.7)
        
        # 添加图例
        plt.legend()
        
        # 设置标题和坐标轴标签
        plt.title(f'Normalized Metric Values Over Time - {filename}')
        plt.xlabel('Time')
        plt.ylabel('Normalized Value')
        
        # 优化 X 轴日期显示
        plt.gcf().autofmt_xdate()
        
        # 保存图表
        output_file = os.path.join(output_folder, f"{os.path.splitext(filename)[0]}.png")
        plt.savefig(output_file)
        plt.close()  # 关闭当前图表，避免内存泄漏

print("所有图表已保存到指定文件夹。")





# import os
# import pandas as pd
# import matplotlib.pyplot as plt
# from sklearn.preprocessing import MinMaxScaler

# # 基本路径
# addr = R'E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\cpu\checkoutservice-1011-1441'

# # 文件夹路径
# abnormal_folder = os.path.join(addr, 'abnormal', 'processed_metrics')
# normal_folder = os.path.join(addr, 'normal', 'processed_metrics')
# output_folder = os.path.join(addr, 'figure')

# # 确保输出文件夹存在
# os.makedirs(output_folder, exist_ok=True)

# # 遍历 abnormal 文件夹中的所有 CSV 文件
# for filename in os.listdir(abnormal_folder):
#     if filename.endswith(".csv"):
#         abnormal_file_path = os.path.join(abnormal_folder, filename)
#         normal_file_path = os.path.join(normal_folder, filename)
        
#         # 检查是否有对应的 normal 文件
#         if not os.path.exists(normal_file_path):
#             print(f"没有找到 {filename} 对应的 normal 文件，跳过此文件。")
#             continue
        
#         # 读取 abnormal 和 normal 文件
#         abnormal_data = pd.read_csv(abnormal_file_path)
#         normal_data = pd.read_csv(normal_file_path)
        
#         # 确保 'TimeUnix' 列被解析为日期时间格式
#         abnormal_data['TimeUnix'] = pd.to_datetime(abnormal_data['TimeUnix'])
#         normal_data['TimeUnix'] = pd.to_datetime(normal_data['TimeUnix'])
        
#         # 准备归一化处理
#         scaler = MinMaxScaler()
        
#         # 选择需要归一化的列
#         metric_columns = [
#             'k8s.pod.cpu.usage',
#             'k8s.pod.cpu_limit_utilization',
#             'k8s.pod.filesystem.usage',
#             'k8s.pod.memory.usage',
#             'k8s.pod.memory_limit_utilization',
#             'k8s.pod.network.errors',
#             'receive_bytes',
#             'transmit_bytes',
#             'k8s.pod.filesystem.utilization'
#         ]

        
#         # 检查两个数据集中是否包含所有所需的列
#         available_columns_abnormal = [col for col in metric_columns if col in abnormal_data.columns]
#         available_columns_normal = [col for col in metric_columns if col in normal_data.columns]
        
#         # 保证两个数据集的度量列一致
#         available_columns = list(set(available_columns_abnormal) & set(available_columns_normal))
        
#         if len(available_columns) == 0:
#             print(f"文件 {filename} 中没有找到任何共同的度量列，跳过此文件。")
#             continue
        
#         # 对 abnormal 和 normal 数据分别进行归一化处理
#         abnormal_data[available_columns] = scaler.fit_transform(abnormal_data[available_columns])
#         normal_data[available_columns] = scaler.transform(normal_data[available_columns])
        
#         # 设置绘图
#         plt.figure(figsize=(15, 10))
        
#         # 定义每个度量的颜色
#         colors = {
#             'k8s.pod.cpu.usage': 'blue',
#             'k8s.pod.cpu_limit_utilization': 'green',
#             'k8s.pod.filesystem.usage': 'red',
#             'k8s.pod.memory.usage': 'cyan',
#             'k8s.pod.memory_limit_utilization': 'magenta',
#             'k8s.pod.network.errors': 'yellow',
#             'receive_bytes': 'black',
#             'transmit_bytes': 'orange',
#             'k8s.pod.filesystem.utilization': 'purple'
#         }
        
#         # 绘制 abnormal 和 normal 数据的折线图
#         for metric in available_columns:
#             color = colors.get(metric, 'gray')  # 如果没有定义颜色，使用灰色
            
#             # 绘制 abnormal 数据
#             plt.plot(abnormal_data['TimeUnix'].values, abnormal_data[metric].to_numpy(), 
#                      label=f'abnormal_{metric}', color=color, linestyle='-', alpha=0.7)
            
#             # 绘制 normal 数据
#             plt.plot(normal_data['TimeUnix'].values, normal_data[metric].to_numpy(), 
#                      label=f'normal_{metric}', color=color, linestyle='--', alpha=0.7)
        
#         # 添加图例
#         plt.legend()
        
#         # 设置标题和坐标轴标签
#         plt.title(f'Normalized Metric Values Over Time - {filename}')
#         plt.xlabel('Time')
#         plt.ylabel('Normalized Value')
        
#         # 优化 X 轴日期显示
#         plt.gcf().autofmt_xdate()
        
#         # 保存图表
#         output_file = os.path.join(output_folder, f"{os.path.splitext(filename)[0]}.png")
#         plt.savefig(output_file)
#         plt.close()  # 关闭当前图表，避免内存泄漏

# print("所有图表已保存到指定文件夹。")
