FROM golang:1.24-alpine AS builder
ENV TZ=Asia/Shanghai
WORKDIR /app

COPY experiments_controller/go.mod experiments_controller/go.sum* ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY experiments_controller /app
RUN go build -o /app/exp

FROM debian:bookworm AS runner
WORKDIR /app

COPY ./manifests/harbor.lab.pj.crt /usr/local/share/ca-certificates/harbor.lab.pj.crt
RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources && apt-get update && \
    apt-get install -y ca-certificates tzdata && \
    chmod 644 /usr/local/share/ca-certificates/harbor.lab.pj.crt && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

COPY --from=builder /app/exp /app/exp
COPY ./algorithms ./algorithms
COPY ./algorithms ./legacy_algorithms
COPY ./experiments_controller/entrypoint.sh ./entrypoint.sh
COPY ./experiments ./experiments
# COPY ./Makefile ./Makefile

RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]