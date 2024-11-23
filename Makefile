# Makefile

ALGO=nezha
BUILDER_IMAGE=$(ALGO)_builder:local
DATA_BUILDER_IMAGE=$(ALGO)_data_builder:local
RUNNER_IMAGE=$(ALGO)_runner:local

.PHONY: all builder data_builder runner clean

all: builder data_builder runner

builder:
	docker run --rm -v /var/lib/dagger --name dagger-engine-v0.14.0 --privileged -v $PWD/manifests/engine.toml:/etc/dagger/engine.toml registry.dagger.io/engine:v0.14.0 