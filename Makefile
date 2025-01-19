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
	kubectl delete jobs --all -n experiment && \
	cd experiments_controller && \
	go run main.go both --port 8082

swagger:
	cd experiments_controller && swag init

gen:
	python scripts/cmd/main.py --algo -d1

jobs:
	kubectl get jobs -n experiment

pods:
	kubectl get pods -n experiment