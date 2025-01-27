PWD := $(shell pwd)


run:
	docker compose down && \
	docker compose up redis mariadb -d && \
	kubectl delete jobs --all -n experiment && \
	cd experiments_controller && \
	go run main.go both --port 8082

swagger:
	swag init \
  	-d ./experiments_controller \
  	--parseDependency \
  	--parseDepth 1

gen:
	python scripts/cmd/main.py --algo -d1

jobs:
	kubectl get jobs -n experiment

pods:
	kubectl get pods -n experiment

build:
	docker build -t 10.10.10.240/library/rcabench:latest -f experiments_controller/Dockerfile .
	docker push 10.10.10.240/library/rcabench:latest
	helm install rcabench ./helm -n experiment

delete:
	helm uninstall rcabench -n experiment