#!/bin/bash

minikube start --nodes 3 \
  --driver=docker \
  --cpus=2 \
  --memory=4g 

# install chaos-mesh
helm repo add chaos-mesh https://charts.chaos-mesh.org
kubectl create ns chaos-mesh
helm install chaos-mesh chaos-mesh/chaos-mesh --namespace chaos-mesh --version 2.8.0
kubectl apply -f manifests/chaos-mesh/rbac.yaml
kubectl create token account-cluster -n default


# install cilium
helm repo add cilium https://helm.cilium.io/
helm install cilium cilium/cilium --version 1.18.4 \
  --namespace kube-system
kubectl apply -f manifests/cilium/metrics.yaml

# install otel-kube-stack
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm install --namespace monitoring --create-namespace -f ./manifests/local-dev/otel-kube-stack.yaml opentelemetry-kube-stack open-telemetry/opentelemetry-kube-stack

# install clickhouse
helm repo add clickstack https://hyperdxio.github.io/helm-charts
helm repo update
helm install --namespace monitoring clickstack clickstack/clickstack -f ./manifests/local-dev/click-stack.yaml


# demo workload
kubectl create ns od
helm install --namespace od otel-demo open-telemetry/opentelemetry-demo -f ./manifests/local-dev/otel-demo-values.yaml --set prometheus.rbac.create=false