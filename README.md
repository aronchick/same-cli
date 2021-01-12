# Getting started
- Find your subscription
```
    # Find your subscription
    az account list -o json | jq '.[] | "\(.name) : \(.id)"'

    # Need to manually pick your subscrption and enter it below.
    export SAME_SUBSCRIPTION_ID='XXXXXXXXXXXXXXXXX'
    az account set --subscription $SAME_SUBSCRIPTION_ID
```

- Install go
- Install kubectl
```
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

- EITHER: Create an AKS Cluster
```
    az aks create --resource-group $SAME_CLUSTER_RG --name myAKSCluster --node-count 0 --enable-addons monitoring --generate-ssh-keys
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

- Install clusterctl
```
curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.12/clusterctl-$GOOS-$GOARCH -o clusterctl
chmod +x ./clusterctl
sudo mv ./clusterctl /usr/local/bin/clusterctl
```

- Set the clusterctl config variables (https://github.com/kubernetes-sigs/cluster-api-provider-azure/blob/master/templates/flavors/README.md)
```
    # Will fail if already there
    mkdir -p ~/.same
    chmod 700 ~/.same

    # Create a service principal
    export SAME_SP_NAME="same_service_principal_$(whoami)"
    export SAME_SP_JSON=`az ad sp create-for-rbac --scope "/subscriptions/$SAME_SUBSCRIPTION_ID/resourceGroups/$SAME_CLUSTER_RG" --role Contributor --sdk-auth`

    export AZURE_ENVIRONMENT="AzurePublicCloud"

    export AZURE_TENANT_ID_B64=$(az account show --subscription=$SAME_SUBSCRIPTION_ID -o json | jq -r '.tenantId' | base64 | tr -d '\n')
    export AZURE_CLIENT_ID=$(echo -n $SAME_SP_JSON | jq -r '.clientId')
    export AZURE_CLIENT_ID_B64=$(echo -n $AZURE_CLIENT_ID | base64 | tr -d '\n')
    export AZURE_CLIENT_SECRET=$(echo -n $SAME_SP_JSON | jq -r '.clientSecret')
    export AZURE_CLIENT_SECRET_B64=$(echo -n $AZURE_CLIENT_SECRET | base64 | tr -d '\n')
    export AZURE_SUBSCRIPTION_ID=$(echo -n $SAME_SUBSCRIPTION_ID | base64 | tr -d '\n')
    export AZURE_SUBSCRIPTION_ID_B64=$(echo -n $SAME_SUBSCRIPTION_ID | base64 | tr -d '\n')
 ```

- OPTIONAL: Init clusterctl
```
    clusterctl init --infrastructure=azure

    # Name of the Azure datacenter location.
    export AZURE_LOCATION="centralus"

    # Select VM types.
    export AZURE_CONTROL_PLANE_MACHINE_TYPE="Standard_D2s_v3"
    export AZURE_NODE_MACHINE_TYPE="Standard_D2s_v3"
```