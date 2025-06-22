# example, pull image to k8s.io namespace, saving time
sudo ctr -n k8s.io image pull quay.io/cilium/cilium-envoy:latest


# install cilium

```bash
helm upgrade cilium cilium/cilium --version 1.17.4 \
   --namespace kube-system -f  cilium-user-values.yaml
```

##  cilium-monitoring

data pipeline:

cilium -> cilium prometheus -> otel collector prome receiver -> clickhouse

```
kubectl apply -f cilium-metrics.yaml
```