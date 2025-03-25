FROM python:3.12.7-bookworm

ENV TZ=Asia/Shanghai

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    gcc \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

RUN pip install uv

WORKDIR /app

COPY . /app/sdk

RUN uv pip install --system --no-cache /app/sdk

CMD ["/bin/bash"]