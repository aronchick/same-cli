#!/usr/bin/env bash
export KUBECONFIG=$HOME/.kube/config
export K3S_CONTEXT="default"
export AKS_CLUSTER="AKSMLProductionCluster"

kubectl port-forward --context=$K3S_CONTEXT svc/ml-pipeline-ui 8080:80 >/dev/null 2>&1 &
kubectl port-forward --context=$AKS_CLUSTER svc/ml-pipeline-ui 9090:80 >/dev/null 2>&1 &