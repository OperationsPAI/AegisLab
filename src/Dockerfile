FROM golang:1.24-bookworm AS builder

ARG GO_BUILD_TAGS

ENV TZ=Asia/Shanghai
WORKDIR /app

RUN apt-get update && apt-get install -y g++ gcc make

COPY src/go.mod src/go.sum* ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

COPY src /app
RUN CGO_ENABLED=1 go build -tags="${GO_BUILD_TAGS}" -ldflags="-s -w" -o /app/exp

FROM debian:bookworm AS runner

ARG INSTALL_CERTS=false

ENV TZ=Asia/Shanghai
WORKDIR /app

RUN if [ "$INSTALL_CERTS" = "true" ]; then \
        echo "Updating sources and installing packages..." && \
        sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources && \
        apt-get update && \
        apt-get install -y ca-certificates tzdata git; \
    fi

COPY ./helm/certs/harbor.lab.pj.crt* /usr/local/share/ca-certificates/

RUN if [ "$INSTALL_CERTS" = "true" ] && [ -f /usr/local/share/ca-certificates/harbor.lab.pj.crt ]; then \
        echo "Updating CA certificates..." && \
        chmod 644 /usr/local/share/ca-certificates/harbor.lab.pj.crt && \
        update-ca-certificates && \
        rm -rf /var/lib/apt/lists/*; \
    fi

COPY --from=builder /app/exp /app/exp
COPY ./src/entrypoint.sh ./entrypoint.sh

RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]