FROM golang:1.24-bookworm AS builder

ARG GO_BUILD_TAGS
ARG GOPROXY="https://proxy.golang.org,direct"

ENV TZ=Asia/Shanghai
WORKDIR /app

RUN apt-get update && apt-get install -y g++ gcc

COPY src/go.mod src/go.sum ./
ENV GOPROXY=$GOPROXY
RUN go mod download

COPY src/ .
RUN CGO_ENABLED=1 go build -tags="${GO_BUILD_TAGS}" -ldflags="-s -w" -o /app/exp
FROM debian:bookworm AS runner

ARG INSTALL_CERTS=false

ENV TZ=Asia/Shanghai
WORKDIR /app

RUN if [ "$INSTALL_CERTS" = "true" ]; then \
        sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources && \
        apt-get update && \
        apt-get install -y ca-certificates tzdata git && \
        if [ -f /usr/local/share/ca-certificates/harbor.lab.pj.crt ]; then \
            chmod 644 /usr/local/share/ca-certificates/harbor.lab.pj.crt && \
            update-ca-certificates; \
        fi && \
        rm -rf /var/lib/apt/lists/*; \
    fi

COPY --from=builder /app/exp /app/exp
COPY ./src/entrypoint.sh ./entrypoint.sh

RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]