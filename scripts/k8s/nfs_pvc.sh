kubectl create ns experiment
kubectl apply -f pv.yaml -n experiment
kubectl apply -f pvc.yaml -n experiment