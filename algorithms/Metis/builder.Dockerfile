FROM python:3.9-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git

RUN git clone --depth 1 https://github.com/CUHK-SE-Group/RCA

RUN cd RCA && pip install --no-cache-dir -r requirements.txt