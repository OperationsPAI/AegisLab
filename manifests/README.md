# Manifests

## Chaos Mesh

### Installation

> [!TIP]
> 📖 **Original Tutorial**: This step is based on the [Chaos Mesh Installation](https://chaos-mesh.org/docs/production-installation-using-helm).

```bash
# Add the Chaos Mesh repository to the Helm repository
helm repo add chaos-mesh https://charts.chaos-mesh.org

helm install chaos-mesh chaos-mesh/chaos-mesh --namespace chaos-mesh
    --version 2.7.2 \
    --set chaosDaemon.runtime=containerd \
    --set chaosDaemon.socketPath=/run/containerd/containerd.sock
```

#### Access Dashboard

```bash
kubectl get svc -n chaos-mesh
NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                 AGE
chaos-daemon                    ClusterIP   None            <none>        31767/TCP,31766/TCP                     30m
chaos-dashboard                 NodePort    10.97.177.177   <none>        2333:31632/TCP,2334:32683/TCP           30m
chaos-mesh-controller-manager   ClusterIP   10.105.196.18   <none>        443/TCP,10081/TCP,10082/TCP,10080/TCP   30m
chaos-mesh-dns-server           ClusterIP   10.96.127.183   <none>        53/UDP,53/TCP,9153/TCP,9288/TCP         30m

kubectl port-forward -n chaos-mesh svc/chaos-dashboard 3500:2333
```

Visit `http://localhost:3500` in your browser to access the Chaos Mesh dashboard.

### Generate Dashboard Token

```bash
# Create the service account and role binding
kubectl apply -f manifests/chaos-mesh/rbac.yaml

# Generate a service account token
kubectl create token account-cluster -n default
```

Copy the generated token and use it to log in to the dashboard. The name of the front end can be written as you like.

## clilum

### installation

```bash
# pull image to k8s.io namespace, saving time
# sudo ctr -n k8s.io image pull quay.io/cilium/cilium-envoy:latest

helm install --namespace kube-system cilium cilium/cilium \
  --version 1.17.4 -f cilium-user-values.yaml \
```

### monitoring

**data pipeline**:

cilium -> cilium prometheus -> otel collector prome receiver -> clickhouse

```bash
kubectl apply -f manifests/cilium/cilium-metrics.yaml
```

## otel-kube-stack

```bash
helm install --namespace monitoring -f otel-kube-stack.yaml\
  opentelemetry-kube-stack open-telemetry/opentelemetry-kube-stack \
```
