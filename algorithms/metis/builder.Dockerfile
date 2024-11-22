FROM python:3.9-slim AS builder

WORKDIR /app

COPY RCA/ /app
ENV PYTHONPATH="/app"
RUN pip install --no-cache-dir -r requirements.txt