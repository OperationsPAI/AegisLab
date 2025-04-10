
# Prepare

```bash
kubectl create secret generic kube-config --from-file=kubeconfig=/home/nn/.kube/config -n experiment
```


# Local debug

```bash
# make sure that Docker is installed
make debug
```

# Deploy

```bash
skaffold run --default-repo=10.10.10.240/library # deploy or upgrade the service to k8s
skaffold debug --default-repo=10.10.10.240/library # debug the service to k8s, if ctrl-c, the helm chart will be uninstalled
```