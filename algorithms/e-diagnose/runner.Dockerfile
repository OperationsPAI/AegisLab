FROM python:3.10-slim AS runner

WORKDIR /app

# 安装运行时依赖
COPY --from=builder:local  /app /app
COPY --from=builder:local  /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages

COPY --from=data_builder:local /app/input.csv /app/input.csv

COPY rca.py .