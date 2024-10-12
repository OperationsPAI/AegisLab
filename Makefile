# Makefile

ALGO=nezha
BUILDER_IMAGE=$(ALGO)_builder:local
DATA_BUILDER_IMAGE=$(ALGO)_data_builder:local
RUNNER_IMAGE=$(ALGO)_runner:local

.PHONY: all builder data_builder runner clean

all: builder data_builder runner

builder:
	docker build -f algorithms/$(ALGO)/builder.Dockerfile -t $(BUILDER_IMAGE) algorithms/$(ALGO)

data_builder: builder
	docker build -f benchmarks/clickhouse/Dockerfile -t $(DATA_BUILDER_IMAGE) benchmarks/clickhouse

runner: data_builder
	docker build -f algorithms/$(ALGO)/runner.Dockerfile --build-arg BUILDER_IMAGE=$(BUILDER_IMAGE) --build-arg DATA_BUILDER_IMAGE=$(DATA_BUILDER_IMAGE)  -t $(RUNNER_IMAGE) algorithms/$(ALGO)

clean:
	docker rmi $(BUILDER_IMAGE) $(DATA_BUILDER_IMAGE) $(RUNNER_IMAGE)