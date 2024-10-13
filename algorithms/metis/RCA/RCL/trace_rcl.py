import pandas as pd

def read_and_merge_data(normal_path, abnormal_path):
    normal_df = pd.read_csv(normal_path)
    abnormal_df = pd.read_csv(abnormal_path)
    merged_df = pd.merge(normal_df, abnormal_df, on=["ServiceName", "SpanName"], suffixes=('_normal', '_abnormal'))
    return merged_df

def calculate_changes(merged_df):
    merged_df['ChangePercentage'] = ((merged_df['MeanDuration_abnormal'] - merged_df['MeanDuration_normal']) / merged_df['MeanDuration_normal']) * 100
    merged_df['ChangeDescription'] = merged_df['ChangePercentage'].apply(
        lambda x: f"increase {abs(x):.2f}%" if x > 0 else f"decrease {abs(x):.2f}%"
    )
    return merged_df

def sort_and_filter_data(merged_df):
    merged_df = merged_df[merged_df['ChangePercentage'] >= 0]
    sorted_df = merged_df.sort_values(by='ChangePercentage', key=abs, ascending=False)
    result_df = sorted_df[['ServiceName', 'SpanName', 'MeanDuration_normal', 'MeanDuration_abnormal', 
                           'ChangeDescription', 'ParentServiceName_abnormal', 'TraceId_abnormal']]
    result_df.to_csv("rcl_output/trace_rcl_results.csv", index=False)
    filtered_df = merged_df[merged_df['ChangePercentage'].abs() > 20].copy()
    return result_df, filtered_df

def weighted_change(filtered_df):
    filtered_df.loc[:, 'WeightedChange'] = filtered_df['ChangePercentage'] * (1 / (filtered_df.index + 1))
    span_count = filtered_df['ServiceName'].value_counts()
    total_spans = len(filtered_df)
    filtered_df.loc[:, 'ProportionalWeightedChange'] = filtered_df.apply(
        lambda row: row['WeightedChange'] * (span_count[row['ServiceName']] / total_spans), axis=1
    )
    result = filtered_df.groupby('ServiceName')['ProportionalWeightedChange'].sum().reset_index()
    result = result.sort_values(by='ProportionalWeightedChange', ascending=False)
    result.to_csv("rcl_output/trace_service_scores.csv", index=False)
    return result

def print_top_spans(file_path):
    df = pd.read_csv(file_path)
    for index, row in df.head(5).iterrows():
        print("trace event - top ", (index + 1))
        print(f"    - span_name: {row['SpanName']}")
        print(f"    - service: {row['ServiceName']} pod")
        print(f"    - parent_service: {row['ParentServiceName_abnormal']} pod")
        print(f"    - normal_duration(per min): {row['MeanDuration_normal'] / 60000000:.2f}s")
        print(f"    - observed_duration: {row['MeanDuration_abnormal'] / 60000000:.2f}s")
        print(f"    - pattern: {row['ChangeDescription']}")
        print(f"    - trace_id: {row['TraceId_abnormal']}")
        print()

def main():
    merged_df = read_and_merge_data("mean_std/normal.csv", "mean_std/abnormal.csv")
    merged_df = calculate_changes(merged_df)
    result_df, filtered_df = sort_and_filter_data(merged_df)
    weighted_change_result = weighted_change(filtered_df)
    print(weighted_change_result)
    print_top_spans("rcl_output/trace_rcl_results.csv")

if __name__ == "__main__":
    main()
