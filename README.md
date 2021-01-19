# User experience

Our goal is to have a user experience that looks like this:
```
$ same create -f https://github.com/contoso/great_nlp_model
```

Assuming you having a vanilla Kubernetes cluster and local credentials, this will:
- Deploy a node pool specific to your SAME description
- Create a namespace for your KF experiment
- Provision Kubeflow to it
- Create data buckets (if necessary) 
- Copy the data into the bucket (if necessary)
- Provision a PV in the namespace and point it at the data
- Push a Kubeflow Pipeline to the Kubeflow
- Allow you to run the pipeline (from either the CLI or the UI)

# Getting your environment variables set up correctly.
```
cp set_env_vars_sample.sh > set_env_vars.sh
```
- Replace everything with an "XXXXXX" with the correct value
- Run the following:

```
. ./set_env_vars.sh
python create_env_file.py
```

# How to build
Just run `make build`

Then run `bin/same`

# Additional installations you probably need to do.

- Install go
- Install kubectl
```
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```
- Install python conveniences (poetry & pre-commit.com):
```
pip install poetry
poetry shell
python -m pip install --upgrade pip
pip install pre-commit
```

- EITHER: Create an AKS Cluster
```
# Set your Resource Group
    export SAME_CLUSTER_RG='XXXXXXXXXXXXXXXXX'
    az aks create --resource-group $SAME_CLUSTER_RG --name same_test_cluster_$(whoami) --node-count 0 --enable-addons monitoring --generate-ssh-keys
```
- OR: Use an existing cluster
```
az aks list --subscription=$SAME_SUBSCRIPTION_ID -o json | jq -r '.[] | "- Cluster Name: \t\(.name) \n  Resource Group: \t\(.resourceGroup)"'
```

- Set Environment Variables for your cluster
```
export SAME_CLUSTER_NAME='XXXXXXXXXXXXXXXXX'
export SAME_CLUSTER_RG='XXXXXXXXXXXXXXXXX'
export SAME_CLUSTER_VERSION=`az aks show -n $SAME_CLUSTER_NAME -g $SAME_CLUSTER_RG -o json | jq -r '.kubernetesVersion'`
```

- Get your credentials:
```
az aks get-credentials -n $SAME_CLUSTER_NAME -g $SAME_CLUSTER_RG
```

# Using SAME
- Be in the same directory as same.yaml



