FROM python:3.6-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git

RUN git clone git clone git@github.com:IntelligentDDS/Nezha.git

RUN cd Nezha && pip install --no-cache-dir -r requirements.txt