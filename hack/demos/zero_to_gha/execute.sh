#!/usr/bin/env bash

########################
# include the magic
########################
. ./demo-magic.sh


########################
# Configure the options
########################

#
# speed at which to simulate typing. bigger num = faster
#
# TYPE_SPEED=20

#
# custom prompt
#
# see http://www.tldp.org/HOWTO/Bash-Prompt-HOWTO/bash-prompt-escape-sequences.html for escape sequences
#
DEMO_PROMPT="${GREEN}➜ ${CYAN} ~ ${WHITE}$ "

# text color
# DEMO_CMD_COLOR=$BLACK

# hide the evidence
DEMO_DIR="/tmp/in-flight-demo"

# put your demo awesomeness here
if [ ! -d "$DEMO_DIR" ]; then
  mkdir $DEMO_DIR
fi

cd $DEMO_DIR
export KUBECONFIG=$HOME/.kube/config
export K3S_CONTEXT="default"
export AKS_CLUSTER="AKSMLProductionCluster"
sudo apt-get install -y python3 python3-pip
pip3 install kfp
sudo update-alternatives --install /usr/bin/python python /usr/bin/python3 1
export PATH=$PATH:$HOME/.local/bin
pkill -f kubectl >/dev/null 2>&1

clear

pe "az aks list"

pe "clear"

pe ""

# Clone the directory
p "git clone git@github.com:Trey-Research-AI-Division/Housing_Model.git"
git clone git@github.com:Trey-Research-AI-Division/Housing_Model.git
p ""

# cd into the directory
p "cd Housing_Model"
cd Housing_Model

DEMO_PROMPT="${GREEN}➜ ${CYAN} ~/Housing_Prices_Model ${WHITE}$ "
# Pretend the user has typed in the command from the website
pe "curl -L0 https://get.sameproject.org | bash -"
pe "clear"

# Install k3s
pe "sudo same installK3s"
pe "clear"

# Switch to K3s behind the scenes
kubectl config use-context $K3S_CONTEXT >/dev/null 2>&1

# Install Kubeflow
pe "same init"
pe "clear"

export EXPERIMENT_NAME="Housing Prices Pipeline"
export RUN_NAME="My_Run"

pe "same program run --experiment-name=\"${EXPERIMENT_NAME}\" --run-name=\"${RUN_NAME}_1\""
p "kubectl port-forward svc/ml-pipeline-ui 8080:80"
kubectl port-forward --context=$K3S_CONTEXT svc/ml-pipeline-ui 8080:80 >/dev/null 2>&1 &

# Change the parameters and run again
pe "same program run --experiment-name=\"${EXPERIMENT_NAME}\" --run-name=\"${RUN_NAME}_2\" --run-param=epochs=42"
p ""

# Change the code and run again
pe "same program run --experiment-name=\"${EXPERIMENT_NAME}\" --run-name=\"${RUN_NAME}_3\""
p ""
pe "clear"

# Satisfied with the results, try it on her cluster
az aks get-credentials -n $AKS_CLUSTER -g SAME-sample-vm_group >/dev/null 2>&1
pe "kubectl config use-context $AKS_CLUSTER"
pe "kubectl config set-context $AKS_CLUSTER --namespace=kubeflow"

# Install Kubeflow
pe "same init"
kubectl port-forward --context=$AKS_CLUSTER svc/ml-pipeline-ui 9090:80 >/dev/null 2>&1 &
pe "clear"

# Run again
pe "same program run --experiment-name=\"STAGING_${EXPERIMENT_NAME}\" --run-name=\"STAGING_${RUN_NAME}_1\""
p "kubectl port-forward svc/ml-pipeline-ui 9090:80"

# Change the parameters and run again
pe "same program run --experiment-name=\"STAGING_${EXPERIMENT_NAME}\" --run-name=\"STAGING_${RUN_NAME}_2\" --run-param=epochs=1337"

# Change the code and run again
pe "same program run --experiment-name=\"STAGING_${EXPERIMENT_NAME}\" --run-name=\"STAGING_${RUN_NAME}_3\""
pe "clear"

# Time to commit
git config --global user.name "John Doe" >/dev/null 2>&1
git config --global user.email "john@doe.org"  >/dev/null 2>&1
pe "git commit -a -m 'Performance looks good, check in the code'"

# Open github actions
pe "git push"

pe "echo \"Go to GitHub now :)\""

# # run command behind
cd ~ && rm -rf $DEMO_DIR

# # enters interactive mode and allows newly typed command to be executed
# cmd

# show a prompt so as not to reveal our true nature after
# the demo has concluded
p ""

# pe "kubectl get deployments -n kubeflow"

# pe "kubectl get namespaces"

# pe "kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80"