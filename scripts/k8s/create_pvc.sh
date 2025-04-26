#!/bin/bash

# Default namespace if not provided
NAMESPACE=${1:-experiment}

kubectl create ns $NAMESPACE
kubectl apply -f algorithm_pv.yaml -n $NAMESPACE
kubectl apply -f dataset_pv.yaml -n $NAMESPACE
kubectl create secret generic kube-config --from-file=config=/home/nn/.kube/config -n $NAMESPACE