import os

import pandas as pd


def read_and_merge_data(normal_path, abnormal_path):
    normal_df = pd.read_csv(normal_path)
    abnormal_df = pd.read_csv(abnormal_path)
    merged_df = pd.merge(normal_df, abnormal_df, on=["ServiceName", "SpanName"], suffixes=('_normal', '_abnormal'))
    return merged_df


def calculate_changes(merged_df):
    merged_df['ChangePercentage'] = (
        (merged_df['MeanDuration_abnormal'] - merged_df['MeanDuration_normal']) / merged_df['MeanDuration_normal']
    ) * 100
    # merged_df['ChangeDescription'] = merged_df['ChangePercentage'].apply(
    #     lambda x: f"increase {abs(x):.2f}%" if x > 0 else f"decrease {abs(x):.2f}%"
    # )
    return merged_df


# def sort_and_filter_data(merged_df):
#     os.makedirs("rcl_output", exist_ok=True)
#     merged_df = merged_df[merged_df['ChangePercentage'] >= 0]
#     sorted_df = merged_df.sort_values(by='ChangePercentage', key=abs, ascending=False)
#     result_df = sorted_df[
#         [
#             'ServiceName',
#             'SpanName',
#             'MeanDuration_normal',
#             'MeanDuration_abnormal',
#             'ChangeDescription',
#             'ParentServiceName_abnormal',
#             'TraceId_abnormal',
#         ]
#     ]
#     result_df.to_csv("rcl_output/trace_rcl_results.csv", index=False)
#     filtered_df = merged_df[merged_df['ChangePercentage'].abs() > 20].copy()
#     return result_df, filtered_df


def sort_and_filter_data(merged_df):
    os.makedirs("rcl_output", exist_ok=True)
    merged_df = merged_df[merged_df['ChangePercentage'] >= 0]
    
    # Grouping the data based on the specified columns
    aggregated_df = merged_df.groupby(
        ['ServiceName', 'SpanName', 'ParentServiceName_abnormal']
    ).agg(
        MeanDuration_normal=('MeanDuration_normal', 'mean'),
        MeanDuration_abnormal=('MeanDuration_abnormal', 'mean'),
        ChangePercentage=('ChangePercentage', 'mean'),
        TraceId_abnormal=('TraceId_abnormal', lambda x: list(x)),
        SpanNum=('TraceId_abnormal', 'size')
    ).reset_index()

    aggregated_df = aggregated_df.sort_values(by='ChangePercentage', key=abs, ascending=False)

    # Adding the ChangeDescription column
    aggregated_df['ChangeDescription'] = aggregated_df['ChangePercentage'].apply(
        lambda x: f"increase {abs(x):.2f}%" if x > 0 else f"decrease {abs(x):.2f}%"
    )

    filtered_df = aggregated_df[aggregated_df['ChangePercentage'].abs() > 5].copy()

    # Dropping the ChangePercentage column
    # filtered_df = filtered_df.drop(columns=['ChangePercentage'])

    filtered_df.loc[:, 'WeightedChange'] = filtered_df['ChangePercentage'] * (1 / (filtered_df.index + 1))
    service_span_count = (filtered_df.groupby('ServiceName')['SpanNum'].sum()).to_dict()
    total_spans = filtered_df['SpanNum'].sum()

    filtered_df.loc[:, 'anomaly_score'] = filtered_df.apply(
        lambda row: row['WeightedChange'] * (service_span_count[row['ServiceName']] / total_spans), axis=1
    )
    result = filtered_df.groupby('ServiceName')['anomaly_score'].sum().reset_index()
    result = result.sort_values(by='anomaly_score', ascending=False)
    # result.to_csv("rcl_output/trace_service_scores.csv", index=False)

    filtered_df.to_csv("rcl_output/trace_rcl_results.csv", index=False)

    return result



def get_top_spans(file_path):
    df = pd.read_csv(file_path).round(3)
    trace_events = []
    for index, row in df.iterrows():
        trace_events.append(
            {
                "top": index + 1,
                "span_name": row['SpanName'],
                "span_num": row['SpanNum'],
                "service": row['ServiceName'],
                "parent_service": row['ParentServiceName_abnormal'],
                "normal_duration_per_min": round(row['MeanDuration_normal'] / 60000000, 3),
                "observed_duration": round(row['MeanDuration_abnormal'] / 60000000, 3),
                "pattern": row['ChangeDescription'],
                "trace_id": row['TraceId_abnormal']
            }
        )
    return trace_events


def main():
    merged_df = read_and_merge_data("mean_std/normal.csv", "mean_std/abnormal.csv")
    merged_df = calculate_changes(merged_df)
    filtered_df = sort_and_filter_data(merged_df)
    get_top_spans("rcl_output/trace_rcl_results.csv")


if __name__ == "__main__":
    main()
