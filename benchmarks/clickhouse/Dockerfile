FROM python:3.10-slim AS data_builder

WORKDIR /app

COPY ./benchmarks/clickhouse/requirements.txt .

COPY ./benchmarks/clickhouse/prepare_inputs.py .

RUN pip install -r requirements.txt

ARG ABNORMAL_START_VAR
ARG ABNORMAL_END_VAR
ARG NORMAL_START_VAR
ARG NORMAL_END_VAR
ENV ABNORMAL_START=${ABNORMAL_START_VAR}
ENV ABNORMAL_END=${ABNORMAL_END_VAR}
ENV NORMAL_START=${NORMAL_START_VAR}
ENV NORMAL_END=${NORMAL_END_VAR}