# Prepare your compilation environment. For example, if you are using some packages that needs to be compiled first, your can use this dockerfile.
FROM 10.10.10.240/library/pandas:latest AS builder
WORKDIR /app