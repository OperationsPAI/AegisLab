FROM python:3.10-slim AS runner

WORKDIR /app

COPY rca.py .