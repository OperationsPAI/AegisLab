# Makefile

BUILDER_IMAGE=builder:local
DATA_BUILDER_IMAGE=data_builder:local
RUNNER_IMAGE=runner:local

.PHONY: all builder data_builder runner clean

all: builder data_builder runner

builder:
	docker build -f algorithms/e-diagnose/builder.Dockerfile -t $(BUILDER_IMAGE) algorithms/e-diagnose

data_builder: builder
	docker build -f benchmarks/clickhouse/Dockerfile -t $(DATA_BUILDER_IMAGE) benchmarks/clickhouse

runner: data_builder
	docker build -f algorithms/e-diagnose/runner.Dockerfile -t $(RUNNER_IMAGE) algorithms/e-diagnose

clean:
	docker rmi $(BUILDER_IMAGE) $(DATA_BUILDER_IMAGE) $(RUNNER_IMAGE)