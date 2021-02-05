# From: https://gist.github.com/aojea/623b20280508ca57b825a40a791f0108

#!/usr/bin/env bash
set -x

convert_kubelet() {
 echo "Converting kubelet on nodes $NODES"
 for n in $NODES; do
  # /var/lib/kubelet/config.yaml
  cat <<EOF >> "/var/lib/kubelet/config.yaml"
featureGates:
  IPv6DualStack: true
EOF
   # Restart kubelet
   docker exec $n systemctl restart kubelet
  echo "Converted kubelet on node $n"
 done
}

convert_cni() {
  # TODO
  echo "Updated CNI"
}

convert_control_plane(){
  # Kubeadm installs apiserver and controller-manager as static pods
  # one the configuration has changed kubelet restart them
  API_CONFIG_FILE="/etc/kubernetes/manifests/kube-apiserver.yaml"
  CM_CONFIG_FILE="/etc/kubernetes/manifests/kube-controller-manager.yaml"

  # update all the control plane nodes
  for n in $CONTROL_PLANE_NODES; do
    echo "Converting control-plane on node $n"
    # kube-apiserver backup config file
    docker exec $n cp ${API_CONFIG_FILE} ${API_CONFIG_FILE}.bak
    # append second service cidr --service-cluster-ip-range=<IPv4 CIDR>,<IPv6 CIDR>
    docker exec $n sed -i -e "s#.*service-cluster-ip-range.*#&\,${SECONDARY_SERVICE_SUBNET}#" ${API_CONFIG_FILE}
    # configure feature gate --feature-gates=IPv6DualStack=true
    docker exec $n sed -i '/service-cluster-ip-range/i\    - --feature-gates=IPv6DualStack=true' ${API_CONFIG_FILE}

    # kube-controller-manager /etc/kubernetes/manifests/kube-controller-manager.yaml
    docker exec $n cp ${CM_CONFIG_FILE} ${CM_CONFIG_FILE}.bak
    # append second service cidr --service-cluster-ip-range=<IPv4 CIDR>,<IPv6 CIDR>
    docker exec $n sed -i -e "s#.*service-cluster-ip-range.*#&\,${SECONDARY_SERVICE_SUBNET}#" ${CM_CONFIG_FILE}
    # append second cluster cidr --cluster-cidr=<IPv4 CIDR>,<IPv6 CIDR>
    docker exec $n sed -i -e "s#.*service-cluster-ip-range.*#&\,${SECONDARY_CLUSTER_SUBNET}#" ${CM_CONFIG_FILE}
    # configure feature gate --feature-gates=IPv6DualStack=true
    docker exec $n sed -i '/service-cluster-ip-range/i\    - --feature-gates=IPv6DualStack=true' ${CM_CONFIG_FILE}
    echo "Finished converting control plane on node $n"
  done

}

usage()
{
    echo "usage: kind_dual_conversion.sh [-n|--name <cluster_name>] [-ss secondary_service_subnet] [-sc secondary_cluster_subnet]"
    echo "Convert a single stack cluster in a dual stack cluster"
}

parse_args()
{
    while [ "$1" != "" ]; do
        case $1 in
            -n | --name )			                              shift
                                          	                CLUSTER_NAME=$1
                                          	;;
            -ss | --secondary-service-subnet )	           	shift
                                          	                SECONDARY_SERVICE_SUBNET=$1
                                          	;;
            -sc | --secondary-cluster-subnet )	           	shift
                                          	                SECONDARY_CLUSTER_SUBNET=$1
                                          	                ;;
            -h | --help )                       usage
                                                exit
                                                ;;
            * )                                 usage
                                                exit 1
        esac
        shift
    done
}

parse_args $*

# Set default values
CLUSTER_NAME=${CLUSTER_NAME:-kind}
DUALSTACK_FEATURE_GATE="IPv6DualStack=true"
SECONDARY_SERVICE_SUBNET=${SECONDARY_SERVICE_SUBNET:-"fd00:10:96::/112"}
SECONDARY_CLUSTER_SUBNET=${SECONDARY_CLUSTER_SUBNET:-"fd00:10:244::/56"}

# KIND nodes
NODES=$(kind get nodes --name ${CLUSTER_NAME})
CONTROL_PLANE_NODES=$(kind get nodes --name ${CLUSTER_NAME} | grep control)
WORKER_NODES=$(kind get nodes --name ${CLUSTER_NAME} | grep worker)

# Create smoke deployment
kubectl apply -f https://gist.githubusercontent.com/aojea/461dd0da7c36b4210737301620ab1a11/raw/b069c90bf08c93d61b3ec710be25911d04420feb/svc-tcp-udp.yaml
if ! kubectl wait --for=condition=ready pods --all --timeout=100s ; then
  echo "smoke test pods are not running"
  kubectl get pods -A -o wide || true
  exit 1
fi

# Start the conversion to dual stack
# TODO revisit order
convert_kubelet
convert_control_plane
convert_cni

# Wait until changes take effect
# 30s is a random value taken based on observations
sleep 30

# Check everything is fine
if ! kubectl wait --for=condition=ready pods --all --timeout=100s ; then
  echo "smoke test pods are not running"
  kubectl get pods -A -o wide || true
  exit 1
fi

# Create dual stack services and check they work
kubectl apply -f https://gist.githubusercontent.com/aojea/90768935ab71cb31950b6a13078a7e92/raw/99ceac308f2b2658c7313198a39fbe24b155ae68/dual-stack.yaml

# Switch services IP families and IPPolicies
# TODO