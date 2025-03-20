PWD := $(shell pwd)
NS 	?= experiment

build:
	skaffold run --default-repo=10.10.10.240/library

run:
	skaffold debug --default-repo=10.10.10.240/library

debug:
	docker compose down && \
	docker compose up redis mariadb -d && \
	kubectl delete jobs --all -n experiment && \
	# sh scripts/k8s/delete_crds.sh ts && \
	cd experiments_controller && go run main.go both --port 8082

swagger:
	swag init \
  	-d ./experiments_controller \
  	--parseDependency \
  	--parseDepth 1

gen:
	python scripts/cmd/main.py --algo -d1

jobs:
	kubectl get jobs -n $(NS)

pods:
	kubectl get pods -n $(NS)

ports:
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &