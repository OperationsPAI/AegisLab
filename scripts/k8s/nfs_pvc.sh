kubectl create ns experiment
kubectl apply -f pv.yaml -n experiment
kubectl apply -f pvc.yaml -n experiment
kubectl create secret generic kube-config --from-file=config=/home/nn/.kube/config -n experiment