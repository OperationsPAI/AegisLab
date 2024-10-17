FROM python:3.10-slim AS data_builder

WORKDIR /app

COPY requirements.txt .

RUN pip install -r requirements.txt

COPY prepare_inputs.py .

RUN python prepare_inputs.py