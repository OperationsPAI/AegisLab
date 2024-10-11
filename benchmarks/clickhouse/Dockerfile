FROM python:3.10-slim AS data_builder

WORKDIR /app

COPY . .

RUN pip install -r requirements.txt

RUN python prepare_inputs.py