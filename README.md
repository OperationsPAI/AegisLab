
# Prepare Env (first time dev/deploy)

| Reource | Name                 | Volume    | Mode       | NFS Path                     | NFS Service     | StorageClass                  | PVC Bind     |
|----------|----------------------|---------|----------------|------------------------------|----------------|-------------------------------|------------------|
| PV       | nfs-shared-pv        | 1024Gi  | ReadWriteMany  | /mnt/data/rcabench_dataset   | 10.26.1.146    | nfs-storage-class             | nfs-shared-pvc   |
| PVC      | nfs-shared-pvc       | 1024Gi  | ReadWriteMany  | —                            | —              | nfs-storage-class             | nfs-shared-pv    |
| PV       | algorithms-data-pv   | 2Gi     | ReadWriteMany  | /mnt/data/rcabench_algo      | 10.26.1.146    | algorithms-data-storage-class | algorithms-data  |
| PVC      | algorithms-data      | 2Gi     | ReadWriteMany  | —                            | —              | algorithms-data-storage-class | algorithms-data-pv |

```bash
# prepare the pv manually, you can replace the config with your own pv. Here we use NFS.
kubectl apply -f scripts/k8s/pv.yaml
```


# Local debug

```bash
# make sure that Docker is installed
make local-debug
```

# Deploy

```bash
skaffold run --default-repo=10.10.10.240/library # deploy or upgrade the service to k8s
skaffold debug --default-repo=10.10.10.240/library # debug the service to k8s, if ctrl-c, the helm chart will be uninstalled

kubectl get pods -n exp # check whether the service is healthy.
```
