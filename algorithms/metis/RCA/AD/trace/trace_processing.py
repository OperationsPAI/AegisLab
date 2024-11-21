import asyncio
import base64
import json
from pathlib import Path

import pandas as pd
from aiomultiprocess import Pool

from ..abstract import Parser
from ..utils.data_utils import TraceDurationCalculator


class TraceParser(Parser):
    def __init__(self, base_dir):
        super().__init__(base_dir)
        self.file_name = "traces.csv"
        self.normal_group_trace_path = self.base_dir / "normal" / self.file_name
        self.abnormal_group_trace_path = self.base_dir / "abnormal" / self.file_name
        self.output_dir = self.base_dir / "trace_ad_output"
        self.output_dir.mkdir(parents=True, exist_ok=True)

    async def parse(self, pool: Pool):
        """Parse trace data and calculate mean and standard deviation of service spans."""
        tasks = [self._process_func(pool, normal=normal) for normal in [True, False]]
        return await asyncio.gather(*tasks)

    async def _process_func(self, pool: Pool, normal=True):
        """Process trace data and calculate mean and standard deviation of service spans."""
        trace_group_path = self.normal_group_trace_path if normal else self.abnormal_group_trace_path
        parsed_csv_path = trace_group_path
        trace_df = await parse_and_save_csv(trace_group_path, parsed_csv_path, pool, batch_size=50000)
        data = await pool.apply(TraceDurationCalculator.calculate, args=(trace_df, normal))
        output_file = self.output_dir / ("normal.csv" if normal else "abnormal.csv")
        await asyncio.to_thread(data.to_csv, output_file, index=False, mode="w")
        return data


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
                    decode_base64(model_data["references"][0]["span_id"]) if model_data["references"] else pd.NA
                ),
            }
        )
    return pd.DataFrame(parsed_data)


async def parse_and_save_csv(input_file, output_parsed_csv, pool: Pool, batch_size=100000) -> pd.DataFrame:
    """Parse the CSV file, handle both old and new formats, and save the parsed data to a new CSV."""

    trace_df = pd.read_csv(input_file, engine="pyarrow")

    if "model" in trace_df.columns:

        batches = [trace_df.iloc[start : start + batch_size] for start in range(0, len(trace_df), batch_size)]

        # Use aiomultiprocessing to process batches in parallel

        results = await pool.map(batch_process, batches)
        parsed_df = pd.concat(results, ignore_index=True)
        parsed_df = resolve_parent_service_name(parsed_df)
        parsed_df["Timestamp"] = pd.to_datetime(parsed_df["Timestamp"])
        parsed_df = parsed_df.sort_values(by="Timestamp")

        await asyncio.to_thread(parsed_df.to_csv, output_parsed_csv, index=False, mode="w")
        return parsed_df
    else:
        return trace_df
