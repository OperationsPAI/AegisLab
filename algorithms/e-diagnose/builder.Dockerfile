FROM 10.10.10.240/library/pandas:latest AS builder

WORKDIR /app

COPY ./algorithms/e-diagnose .

COPY ./experiments/run_exp.py .



