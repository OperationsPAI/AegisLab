PWD := $(shell pwd)

builder:
	docker run -itd \
		-v /var/lib/dagger \
		--name dagger-engine-v0.14.0 \
		--privileged \
		-v $(PWD)/manifests/engine.toml:/etc/dagger/engine.toml \
		registry.dagger.io/engine:v0.14.0

run:
	docker compose down && \
	docker compose up redis -d && \
	cd experiments_controller && \
	go run main.go both --port 8082

gen:
	python scripts/cmd/main.py -d


bench:
	docker buildx build -t 10.10.10.240/library/clickhouse_dataset:latest benchmarks/clickhouse --push