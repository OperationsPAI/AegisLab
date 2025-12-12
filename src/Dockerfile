FROM golang:1.24-alpine AS builder
ENV TZ=Asia/Shanghai
WORKDIR /app

COPY src/go.mod src/go.sum* ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY src /app
RUN go build -o /app/exp

FROM debian:bookworm AS runner
ENV TZ=Asia/Shanghai
WORKDIR /app

COPY ./helm/certs/harbor.lab.pj.crt /usr/local/share/ca-certificates/harbor.lab.pj.crt
# Install ca-certificates first with default sources, then switch to mirror
RUN apt-get update && \
    apt-get install -y ca-certificates tzdata git && \
    sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources && \
    chmod 644 /usr/local/share/ca-certificates/harbor.lab.pj.crt && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/exp /app/exp
COPY ./src/entrypoint.sh ./entrypoint.sh

RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]