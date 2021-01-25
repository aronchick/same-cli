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
cp set_env_vars_sample.sh set_env_vars.sh
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

- Set your subscription ID
```
az login
```
# Set your Resource Group
```
az account list -o json | jq '.[] | "\(.name) : \(.id)"'
export SAME_SUBSCRIPTION_ID='XXXXXXXXXXXXXXXXX'
az account set --subscription $SAME_SUBSCRIPTION_ID
```

- EITHER: Create an AKS Cluster

# Select a resource group from the above list
```
export SAME_CLUSTER_RG='XXXXXXXXXXXXXXXXX'
export SAME_CLUSTER_NAME="same_test_cluster_$(whoami)"
az aks create --resource-group $SAME_CLUSTER_RG --name $SAME_CLUSTER_NAME --node-count 0 --enable-addons monitoring --generate-ssh-keys
```

- OR: Use an existing cluster
```
az login
az aks list --subscription=$SAME_SUBSCRIPTION_ID -o json | jq -r '.[] | "- Cluster Name: \t\(.name) \n  Resource Group: \t\(.resourceGroup)"'
export SAME_CLUSTER_NAME='XXXXXXXXXXXXXXXXX'
export SAME_CLUSTER_RG='XXXXXXXXXXXXXXXXX'
```

- INSTALL TERRAFORM:
```
curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt-get update && sudo apt-get install terraform
```

- Go into `cmd/infrastructure/azure` and type the following commands:
```
export SAME_PREFIX="same"
export SAME_LOCATION="west europe"

terraform init
terraform plan -var "prefix=$(SAME_PREFIX)" -var "location=$(SAME_LOCATION)"
terraform plan -var "prefix=$(SAME_PREFIX)" -var "location=$(SAME_LOCATION)"
```

- Set Environment Variables for your cluster
```
export SAME_CLUSTER_VERSION=`az aks show -n $SAME_CLUSTER_NAME -g $SAME_CLUSTER_RG -o json | jq -r '.kubernetesVersion'`
```

- Get your credentials:
```
az aks get-credentials -n $SAME_CLUSTER_NAME -g $SAME_CLUSTER_RG
```

# Using SAME
- Be in the same directory as same.yaml

# Goal vs non-goals.

```
Todo: 

* This section could be moved somewhere to the top once Once this section is mature it can sit somewhere at the top of this read-me.
* This seciton will grow as this CLI will enrich with feature.
* Key contributors please feel free to add or delete antyhing for this section. (It is created as placeholder to grow)

```

Goal of this work to build the Kubernetes command-line tool, SAME CLI, for allowing users reproducing machine learning pipelines for their kubernetes cluster. Kubernetes enable users to achieve great things but the key aspect which is missing is the gap between the installation steps to till your kubernetes cluster is ready for the use case. This tool not only takes those early installation \ set-up process but also build reproducible pipelines for the end user and enable them to focus towards Machine Learning model implementations.

Non-Goals: This tool donot intend replace any of the existing eco-systems, this tool main intent as exaplined above it to focus in making ease of use simpler for monolithic \ repetetive machine learning pipelines simpler.
