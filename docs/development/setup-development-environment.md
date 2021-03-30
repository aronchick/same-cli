# Getting your environment variables set up correctly.
- Update your system.
```
sudo apt-get update && sudo apt-get upgrade -y
```
- Download go 1.16

  We use Go 1.16 (https://golang.org/dl/) so that we can utilize some of the latest language features:
  * embed package for static content
  * module-aware builds by default
  * no automatic changes to go.mod and go.sum
  
  
  You'll probably use the following to install go:
```
wget https://golang.org/dl/go1.16.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.16.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.profile
go -v
```

- Install Python 3.9

  https://www.python.org/downloads/

- Log into Azure
```
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
az login
```

- Set your Resource Group
```
sudo apt install jq -y
az account list -o json | jq '.[] | "\(.name) : \(.id)"'
```
Pick your subscription:
```
export AZURE_SUBSCRIPTION_ID=XXXXXXXXXXXXXXXXX
az account set --subscription $AZURE_SUBSCRIPTION_ID
```

- Install kubectl
```
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```
For convenience `kubectl` can be installed via `az aks install-cli`

- Install python conveniences (poetry & pre-commit.com):
```
sudo apt install software-properties-common -y 
sudo add-apt-repository ppa:deadsnakes/ppa -y
sudo apt update -y
sudo apt install python3.9 -y
sudo update-alternatives --install /usr/bin/python python /usr/bin/python3.9 1
sudo apt install python3.9-distutils
curl -LO "https://bootstrap.pypa.io/get-pip.py" > get-pip.py
python3 get-pip.py
echo "export PATH=$PATH:~/.local/bin" >> ~/.bashrc
pip3 install --upgrade pip setuptools distlib keyrings.alt poetry
```
- Clone the repo
```
git clone git@github.com:azure-octo/same-cli.git
```
- Install the poetry venv
```
cd same-cli
python -m poetry shell
python -m pip install --upgrade pip
pip install pre-commit
pre-commit install
```

# How to build
Just run `make build`
```
# The below will go away soon
mkdir ~/.same
touch ~/.same/config.yaml
echo "activepipeline: nil" >> ~/.same/config.yaml
```
- Create your first Kubeflow cluster (this will create a local k3s cluster and then install kubeflow on it)
```
sudo bin/same installK3s
bin/same init
```

To create a sample pipeline:
```
bin/same program create -f same.yaml
```

(Note that `same program create` only creates the pipeline, it doesn't run it.)

To access the ml pipeline UI:
```
kubectl port-forward svc/ml-pipeline-ui 8080:80
```
Then go to [http://localhost:8080](http://localhost:8080)

(Note the sample pipeline doesn't current work on k3s.)


- Run the full tests

```
make test
```
# Goal vs non-goals.

```
Todo: 

* This section could be moved somewhere to the top once Once this section is mature it can sit somewhere at the top of this read-me.
* This seciton will grow as this CLI will enrich with feature.
* Key contributors please feel free to add or delete antyhing for this section. (It is created as placeholder to grow)

```

Goal of this work to build the Kubernetes command-line tool, SAME CLI, for allowing users reproducing machine learning pipelines for their kubernetes cluster. Kubernetes enable users to achieve great things but the key aspect which is missing is the gap between the installation steps to till your kubernetes cluster is ready for the use case. This tool not only takes those early installation \ set-up process but also build reproducible pipelines for the end user and enable them to focus towards Machine Learning model implementations.

Non-Goals: This tool donot intend replace any of the existing eco-systems, this tool main intent as exaplined above it to focus in making ease of use simpler for monolithic \ repetetive machine learning pipelines simpler.

