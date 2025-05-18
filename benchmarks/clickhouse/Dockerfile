# https://github.com/LGU-SE-Internal/rcabench-platform/blob/main/cli/prepare_inputs.py
FROM 10.10.10.240/library/rcabench-platform:5588fef

WORKDIR /app

COPY benchmarks/clickhouse/entrypoint.sh /entrypoint.sh

ARG ABNORMAL_START_VAR
ARG ABNORMAL_END_VAR
ARG NORMAL_START_VAR
ARG NORMAL_END_VAR
ENV ABNORMAL_START=${ABNORMAL_START_VAR}
ENV ABNORMAL_END=${ABNORMAL_END_VAR}
ENV NORMAL_START=${NORMAL_START_VAR}
ENV NORMAL_END=${NORMAL_END_VAR}

ENTRYPOINT ["/entrypoint.sh"]
