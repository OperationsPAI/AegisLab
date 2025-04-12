FROM golang:1.23 AS builder
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY experiments_controller /app
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy
RUN go build -o /app/exp 

FROM debian:latest AS runner
WORKDIR /app
COPY ./algorithms ./algorithms
COPY ./algorithms ./legacy_algorithms
COPY ./experiments_controller/entrypoint.sh ./entrypoint.sh
COPY ./benchmarks ./benchmarks
COPY ./experiments ./experiments
COPY ./manifests/harbor.lab.pj.crt /usr/local/share/ca-certificates/harbor.lab.pj.crt
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*
RUN chmod 644 /usr/local/share/ca-certificates/harbor.lab.pj.crt && \
    update-ca-certificates
COPY --from=builder /app/exp /app/exp
RUN chmod +x ./entrypoint.sh