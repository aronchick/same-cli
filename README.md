# Getting started
- Find your subscription
```
    # Find your subscription
    az account list -o json | jq '.[] | "\(.name) : \(.id)"'

    # Need to manually pick your subscrption and enter it below.
    export SAME_SUBSCRIPTION_ID='XXXXXXXXXXXXXXXXX'
    az account set --subscription $SAME_SUBSCRIPTION_ID
```

- Have a Kubernetes cluster hosted on AKS (there's nothing SPECIFIC about SAME for Azure only, but we're just getting started)
```
az aks list --subscription=$SAME_SUBSCRIPTION_ID -o json | jq -r '.[] | "\(.name) : \(.resourceGroup)"'
export CLUSTER_NAME='XXXXXXXXXXXXXXXXX'
export CLUSTER_RESOURCE_GROUP='XXXXXXXXXXXXXXXXX'
export CLUSTER_VERSION=`az aks show -n $CLUSTER_NAME -o json | jq -r '.kubernetesVersion'`
```
- Make sure you have local credentials for a Kubernetes cluster
- Install go
- Install kubectl
```
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```
- Install clusterctl
```
curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.6/clusterctl-linux-amd64 -o clusterctl
chmod +x ./clusterctl
sudo mv ./clusterctl /usr/local/bin/clusterctl
```
- Set the clusterctl config variables (https://github.com/kubernetes-sigs/cluster-api-provider-azure/blob/master/templates/flavors/README.md)
```
    # Will fail if already there
    mkdir ~/.same
    chmod 700 ~/.same

    # Create a service principal
    export SAME_SP_NAME="same_service_principal_$(whoami)"
    export SAME_SP_JSON=`az ad sp create-for-rbac --scope "/subscriptions/$SAME_SUBSCRIPTION_ID/resourceGroups/$CLUSTER_RESOURCE_GROUP" --role Contributor --sdk-auth`

    export AZURE_ENVIRONMENT="AzurePublicCloud"

    export AZURE_TENANT_ID_B64="`az account show --subscription=$SAME_SUBSCRIPTION_ID -o json | jq -r '.tenantId' | base64` | tr -d '\n'"
    export AZURE_CLIENT_ID=`echo $SAME_SP_JSON | jq -r '.clientId'`
    export AZURE_CLIENT_ID_B64="`echo $AZURE_CLIENT_ID | base64` | tr -d '\n'"
    export AZURE_CLIENT_SECRET=`echo $SAME_SP_JSON | jq -r '.clientId'`
    export AZURE_CLIENT_SECRET_B64="`echo $AZURE_CLIENT_SECRET | base64 | tr -d '\n'`"
    export AZURE_SUBSCRIPTION_ID="`echo $SAME_SUBSCRIPTION_ID | base64 | tr -d '\n'`"
    export AZURE_SUBSCRIPTION_ID_B64="`echo $SAME_SUBSCRIPTION_ID | base64 | tr -d '\n'`"
 ```

- Init clusterctl
```
    clusterctl init --infrastructure=azure

    # Name of the Azure datacenter location.
    export AZURE_LOCATION="centralus"

    # Select VM types.
    export AZURE_CONTROL_PLANE_MACHINE_TYPE="Standard_D2s_v3"
    export AZURE_NODE_MACHINE_TYPE="Standard_D2s_v3"
```
- 