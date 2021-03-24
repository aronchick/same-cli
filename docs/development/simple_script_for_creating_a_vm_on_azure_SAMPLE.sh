#/bin/bash
set -e
export PREFIX=<YOUR_PREFIX>
export VMNAME=PREFIX_clean-vm-011
export RG=<YOUR_RESOURCE_GROUP>
export ADMIN_USER=$(whoami)
export VNET=<YOUR_VNET>
export SUBSCRIPTION=<YOUR_SUBSCRIPTION>

az vm create \
--resource-group $RG \
--image UbuntuLTS \
--admin-username $ADMIN_USER \
--vnet-name $VNET \
--size Standard_D8s_v3 \
--ssh-key-values @~/.ssh/id_rsa.pub \
--location eastus2 \
--subscription $SUBSCRIPTION \
--subnet default \
--name $VMNAME

# Make sure you have your pub key in ~/.ssh/authorized_keys or this will stop access to the machiene
scp ~/.ssh/* $VMNAME:/home/$ADMIN_USER/.ssh

