FROM golang:1.23 as builder

ENV TZ=Asia/Shanghai

WORKDIR /app/experiments_controller

COPY go.mod go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy

COPY . .

RUN go build -o /app/experiments_controller/experiments_controller .
