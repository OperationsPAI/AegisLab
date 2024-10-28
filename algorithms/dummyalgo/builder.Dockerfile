# Prepare your compilation environment. For example, if you are using some packages that needs to be compiled first, your can use this dockerfile.
FROM python:3.10-slim AS builder
WORKDIR /app