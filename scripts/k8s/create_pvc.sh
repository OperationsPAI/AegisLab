kubectl create ns experiment
kubectl apply -f algorithm_pv.yaml -n experiment
kubectl apply -f dataset_pv.yaml -n experiment
kubectl create secret generic kube-config --from-file=config=/home/nn/.kube/config -n experiment