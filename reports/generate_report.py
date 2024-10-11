# scripts/generate_report.py
import json
import sys
import os

def generate_markdown_report(report_json_path, report_md_path):
    with open(report_json_path, 'r') as f:
        metrics = json.load(f)
    
    report_content = f"""
# 根因定位算法评估报告

**准确率**: {metrics['accuracy']:.2f}

**时间消耗**: {metrics['time']:.2f} 秒

**内存使用**: {metrics['memory']:.2f} MB
"""
    
    with open(report_md_path, 'w') as f:
        f.write(report_content)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: generate_report.py <input_json> <output_md>")
        sys.exit(1)
    
    input_json = sys.argv[1]
    output_md = sys.argv[2]
    
    generate_markdown_report(input_json, output_md)