#!/bin/bash
set -e

cd ~

# Update all packages and git
sudo add-apt-repository ppa:git-core/ppa -y
sudo apt-get -y update
sudo apt-get -y upgrade
sudo apt-get install git -y
git --version

# Install go (could be done in a GitHub Action)
wget https://golang.org/dl/go1.16.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.16.2.linux-amd64.tar.gz
export PATH="$PATH:/usr/local/go/bin"
echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.bashrc
rm go1.16.2.linux-amd64.tar.gz

# Install Kubectl
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl
sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubectl

# Install make
sudo apt-get install -y make
sudo apt-get install -y gcc

# Install Azure CLI
sudo apt-get update
sudo apt-get install -y ca-certificates curl apt-transport-https lsb-release gnupg
curl -sL https://packages.microsoft.com/keys/microsoft.asc |
gpg --dearmor |
sudo tee /etc/apt/trusted.gpg.d/microsoft.gpg > /dev/null
AZ_REPO=$(lsb_release -cs)
echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $AZ_REPO main" |
sudo tee /etc/apt/sources.list.d/azure-cli.list
sudo apt-get update
sudo apt-get install -y azure-cli

cat << EOF
The installation is finished. Run the following in your shell

echo "export KUBECONFIG=\$HOME/.kube/config" >> $HOME/.bashrc
source ~/.bashrc
make build
sudo bin/same installK3s
SAME_TARGET=local bin/same init # Installs k3s and default pipeline
EOF
