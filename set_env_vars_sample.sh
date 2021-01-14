export SAME_SUBSCRIPTION_ID='XXXXXXXXXXXXXXXXXX'
az account set --subscription $SAME_SUBSCRIPTION_ID
export SAME_CLUSTER_NAME='XXXXXXXXXXXXXXXXXX'
export SAME_CLUSTER_RG='XXXXXXXXXXXXXXXXXX'
export SAME_CLUSTER_VERSION=`az aks show -n $SAME_CLUSTER_NAME -g $SAME_CLUSTER_RG -o json | jq -r '.kubernetesVersion'`
az aks get-credentials -n $SAME_CLUSTER_NAME -g $SAME_CLUSTER_RG
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
export AZURE_LOCATION="centralus"
export AZURE_CONTROL_PLANE_MACHINE_TYPE="Standard_D2s_v3"
export AZURE_NODE_MACHINE_TYPE="Standard_D2s_v3"

# To write this to your .env file
# Execute python3 create_env_file.py