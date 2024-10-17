import asyncio
import base64
import json
from pathlib import Path

import pandas as pd
from aiomultiprocess import Pool
from .calculate import calculate_abnormal, calculate_normal


def decode_base64(data):
    """Decode base64 encoded data."""
    return base64.b64decode(data).hex()


def resolve_parent_service_name(parsed_df: pd.DataFrame) -> pd.DataFrame:
    """Resolve ParentServiceName by looking up ServiceName for matching ParentSpanId."""
    span_id_to_service = parsed_df.set_index("SpanId")["ServiceName"].to_dict()
    parsed_df["ParentServiceName"] = parsed_df["ParentSpanId"].map(span_id_to_service)
    mask = parsed_df["ParentSpanId"].notna() & parsed_df["ParentServiceName"].isna()
    parsed_df.loc[mask, "ParentSpanId"] = pd.NA
    return parsed_df


async def batch_process(batch: pd.DataFrame):
    parsed_data = []
    for _, row in batch.iterrows():
        model_data = json.loads(row["model"])
        parsed_data.append(
            {
                "Timestamp": row["timestamp"],
                "TraceId": decode_base64(model_data["trace_id"]),
                "SpanId": decode_base64(model_data["span_id"]),
                "SpanName": model_data["operation_name"],
                "ServiceName": model_data["process"]["service_name"],
                "Duration": model_data["duration"],
                "ParentSpanId": (
                    decode_base64(model_data["references"][0]["span_id"])
                    if model_data["references"]
                    else pd.NA
                ),
            }
        )
    return pd.DataFrame(parsed_data)


async def parse_and_save_csv(
    input_file, output_parsed_csv, pool: Pool, batch_size=100000
) -> pd.DataFrame:
    """Parse the CSV file, handle both old and new formats, and save the parsed data to a new CSV."""

    trace_df = pd.read_csv(input_file, engine="pyarrow")

    if "model" in trace_df.columns:
        batches = [
            trace_df.iloc[start : start + batch_size]
            for start in range(0, len(trace_df), batch_size)
        ]

        # Use aiomultiprocessing to process batches in parallel

        results = await pool.map(batch_process, batches)
        parsed_df = pd.concat(results, ignore_index=True)
        parsed_df = resolve_parent_service_name(parsed_df)
        parsed_df["Timestamp"] = pd.to_datetime(parsed_df["Timestamp"])
        parsed_df = parsed_df.sort_values(by="Timestamp")

        await asyncio.to_thread(
            parsed_df.to_csv, output_parsed_csv, index=False, mode="w"
        )
        return parsed_df
    else:
        return trace_df


async def process_func(
    trace_group_path: Path, output_dir: Path, pool: Pool, normal=True
):
    """Process trace data and calculate mean and standard deviation of service spans."""
    parsed_csv_path = trace_group_path  # overwrite the original file
    trace_df = await parse_and_save_csv(
        trace_group_path, parsed_csv_path, pool, batch_size=50000
    )
    if normal:
        output_file = Path(output_dir) / "normal.csv"
        data = calculate_normal(trace_df)
    else:
        output_file = Path(output_dir) / "abnormal.csv"
        data = await calculate_abnormal(trace_df, pool)
    await asyncio.to_thread(data.to_csv, output_file, index=False, mode="w")
    return data


async def parse(base_dir, pool: Pool, file_name="traces.csv"):
    """Parse trace data and calculate mean and standard deviation of service spans."""
    base_dir = Path(base_dir)
    normal_group_trace_path = base_dir / "normal" / file_name
    abnormal_group_trace_path = base_dir / "abnormal" / file_name
    output_dir = base_dir / "trace_ad_output"
    output_dir.mkdir(parents=True, exist_ok=True)
    tasks = [
        process_func(normal_group_trace_path, output_dir, pool, normal=True),
        process_func(abnormal_group_trace_path, output_dir, pool, normal=False),
    ]

    return await asyncio.gather(*tasks)
