FROM 10.10.10.240/library/metis:50fd547 AS builder

WORKDIR /app
COPY ./experiments/entrypoint.sh /