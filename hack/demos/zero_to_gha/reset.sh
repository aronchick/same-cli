#!/bin/bash
kubectl config use-context AKSMLProductionCluster
kubectl delete namespace kubeflow
/usr/local/bin/k3s-uninstall.sh
pkill -f kubectl 