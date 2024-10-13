ARG BUILDER_IMAGE=builder:local
ARG DATA_BUILDER_IMAGE=data_builder:local

FROM ${BUILDER_IMAGE} AS base
FROM ${DATA_BUILDER_IMAGE} AS data_builder
FROM python:3.9-slim AS runner


WORKDIR /app

COPY --from=base  /app /app
COPY --from=base  /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages
COPY --from=data_builder /app/input.csv /app/input.csv

COPY rca.py .