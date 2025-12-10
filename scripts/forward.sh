#!/bin/bash

NAMESPACE=exp

services=$(kubectl get svc -n $NAMESPACE -o json \
  | jq -r '.items[] | .metadata.name + ":" + (.spec.ports[0].port|tostring)')

for svc in $services; do
  name="${svc%%:*}"
  port="${svc##*:}"
  echo "Forwarding $name $port -> localhost:$port"
  kubectl port-forward -n $NAMESPACE svc/$name $port:$port >/dev/null 2>&1 &
done

echo "All forwards started."