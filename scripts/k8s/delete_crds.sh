NAMESPACE=$1
for CRD in $(kubectl get crd -o name | grep "chaos-mesh.org" | cut -d'/' -f2); do
  echo "Deleting all $CRD in namespace $NAMESPACE"
  kubectl delete $CRD -n $NAMESPACE --all
done