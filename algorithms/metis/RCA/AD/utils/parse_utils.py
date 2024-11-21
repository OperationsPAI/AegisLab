import json
import re
from datetime import datetime
from typing import List

import pandas as pd

EXTRACT_REGEX = re.compile(r"(INFO|ERROR|WARN)\s+(.*)\s+TraceID:.*SpanID:\s+ (.*)")
ACCEPT_LOG_LEVELS = ("INFO", "ERROR", "WARN")


def parse_timestamp(timestamp, factor=1, fmt="%H:%M:%S.%f"):
    return datetime.fromtimestamp(timestamp / factor).strftime(fmt)[:-3]


def parse_iso_timestamp(iso_str):
    try:
        return datetime.fromisoformat(iso_str).strftime("%H:%M:%S.%f")[:-3]
    except ValueError:
        return iso_str


# TrainTicket Node.js ts-ui-dashboard
def handle_ts_js(log_data, level):
    """
    {"level":"info","ts":1726140979.7913857,"logger":"http.log.access.log0","msg":"handled request","request":{"method":"GET","host":"10.10.10.220:30080","uri":"/api/v1/contactservice/contacts/account/4d2a46c7-71cb-4cf1-b5bb-b68406d9da6f"}}
    """
    timestamp = parse_timestamp(log_data["ts"])
    return {
        "Body": f"{timestamp} {level:<5} {log_data['request']['method']} {log_data['request']['uri']} {log_data['status']}",
        "message": f"{log_data['request']['method']} {log_data['request']['uri']} {log_data['status']}",
        "log_level": level,
    }


# OnlineBoutique Node.js currencyservice paymentservice
def handle_hipster_js(log_data, level):
    """
    {"severity":"info","time":1728106908868,"pid":1,"hostname":"currencyservice-76756b6d64-jqz8v","name":"currencyservice-server","message":"conversion request successful"}
    """
    timestamp = parse_timestamp(log_data["time"], 1000)
    return {
        "Body": f"{timestamp} {level:<5} {log_data['name']} {log_data['message']}",
        "message": f"{log_data['name']} {log_data['message']}",
        "log_level": level,
    }


# OnlineBoutique Java adservice
def handle_hipster_java(log_data, level):
    """
    {"instant":{"epochSecond":1728107136,"nanoOfSecond":724744806},"level":"INFO","loggerName":"hipstershop.AdService","message":"received ad request (context_words=[kitchen])","time":"2024-10-05T05:45:36.724Z"}
    """
    timestamp = parse_iso_timestamp(log_data["time"].replace("Z", "+00:00"))
    return {
        "Body": f"{timestamp} {level:<5} {log_data['message']}",
        "message": f"{log_data['message']}",
        "log_level": level,
    }


# OnlineBoutique Python emailservice recommendationservice
def handle_hipster_python(log_data, level):
    """
    {"timestamp": 1728106897.9604454, "severity": "INFO", "name": "emailservice-server", "message": "A request to send order confirmation email to kayla42@example.net has been received.", "taskName": null}
    """
    timestamp = parse_timestamp(log_data["timestamp"])
    return {
        "Body": f"{timestamp} {level:<5} {log_data['name']} {log_data['message']}",
        "message": f"{log_data['name']} {log_data['message']}",
        "log_level": level,
    }


# OnlineBoutique go frontend shippingservice checkoutservice
def handle_hipster_go(log_data, level):
    """
    {"http.req.method":"GET","http.req.path":"/product/66VCHSJNUP","message":"request started","severity":"debug","timestamp":"2024-10-05T05:40:43.813174664Z"}
    {"message":"[GetQuote] received request","severity":"info","timestamp":"2024-10-05T05:47:12.650259663Z"}
    """

    timestamp = parse_iso_timestamp(log_data["timestamp"][:26] + "+00:00")

    if "http.req.method" in log_data:
        return {
            "Body": f"{timestamp} {level:<5} {log_data['http.req.method']} {log_data['http.req.path']}",
            "message": f"{log_data['http.req.method']} {log_data['http.req.path']}",
            "log_level": level,
        }
    return {
        "Body": f"{timestamp} {level:<5} {log_data['message']}",
        "message": f"{log_data['message']}",
        "log_level": level,
    }


def handle_fallback(log):
    match = EXTRACT_REGEX.search(log)
    if match:
        return {
            "Body": log,
            "message": f"{match.group(2)} {match.group(3)}",
            "log_level": match.group(1).upper(),
        }
    return {
        "Body": log,
        "message": pd.NA,
        "log_level": pd.NA,
    }


def extract_log(df):
    """Extract log level and log message from log."""
    log = df.Body
    # TrainTicket
    if level := df.get("SeverityNumber"):
        if level == 1:
            level = "TRACE"
        elif level == 5:
            level = "DEBUG"
        elif level == 9:
            level = "INFO"
        elif level == 13:
            level = "WARN"
        elif level == 17:
            level = "ERROR"
        else:
            level = pd.NA
        try:
            log_data = json.loads(log)
            _, message, _ = handle_ts_js(log_data, level).values()
            log = message
        except:
            pass
        return {
            "Body": log,
            "message": log,
            "log_level": level,
        }
    try:
        log_data = json.loads(log)
        level = log_data.get("level", log_data.get("severity", pd.NA))
        if level:
            level = level.upper()
            if level not in ACCEPT_LOG_LEVELS:
                level = pd.NA

        if log_data.get("ts"):
            return handle_ts_js(log_data, level)
        elif isinstance(log_data.get("time"), (int, float)):
            return handle_hipster_js(log_data, level)
        elif isinstance(log_data.get("time"), str):
            return handle_hipster_java(log_data, level)
        elif isinstance(log_data.get("timestamp"), (int, float)):
            return handle_hipster_python(log_data, level)
        elif isinstance(log_data.get("timestamp"), str):
            return handle_hipster_go(log_data, level)
    except Exception:
        return handle_fallback(log)

    return {
        "Body": log,
        "message": pd.NA,
        "log_level": pd.NA,
    }
