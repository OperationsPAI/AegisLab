FROM 10.10.10.240/library/sdk_python:latest

ENV TZ=Asia/Shanghai

WORKDIR /app

RUN apt-get update && apt-get install -y curl tar gzip

ENV HELM_VERSION="v3.14.1"
RUN curl -L https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar -xz && \
    mv linux-amd64/helm /usr/local/bin/ && \
    rm -rf linux-amd64

COPY main.py . 

CMD [ "/bin/bash" ]