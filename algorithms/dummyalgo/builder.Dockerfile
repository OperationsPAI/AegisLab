# Prepare your compilation environment. For example, if you are using some packages that needs to be compiled first, your can use this dockerfile.
FROM python:3.12.7-bookworm AS builder

WORKDIR /app

COPY ./algorithms/dummyalgo .

COPY ./experiments/run_exp.py .

COPY ./experiments/entrypoint.sh /