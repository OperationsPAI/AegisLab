PWD := $(shell pwd)
NS 	?= experiment

build:
	skaffold run --default-repo=10.10.10.240/library

run:
	skaffold debug --default-repo=10.10.10.240/library

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