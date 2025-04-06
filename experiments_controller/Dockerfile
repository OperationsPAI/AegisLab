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
COPY --from=builder /app/exp /app/exp
RUN chmod +x ./entrypoint.sh