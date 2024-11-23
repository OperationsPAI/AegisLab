FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy

COPY . .

RUN go build -o /app/experiments_controller .
