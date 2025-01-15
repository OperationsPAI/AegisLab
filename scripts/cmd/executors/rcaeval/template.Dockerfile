FROM 10.10.10.240/library/rcaeval:1.0

WORKDIR /app

COPY ./algorithms/nsigma .

COPY ./experiments/run_exp.py .