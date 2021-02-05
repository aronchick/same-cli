# Setup SAME development environment

This document helps you get started developing SAME. If you find any problem while following this guide, please create a Pull Request to update this document. (from https://github.com/dapr/dapr/edit/master/docs/development/setup-dapr-development-env.md)

## Docker environment

1. Install [Docker](https://docs.docker.com/install/)
    ```bash
    sudo apt-get update
    sudo apt-get install -y \
         apt-transport-https \
         ca-certificates \
         curl \
         gnupg-agent \
         software-properties-common
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository \
      "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) \
      stable"
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io
    sudo docker run hello-world
    ```

    > For Linux, you'll have to configure docker to run without sudo for this to work, because of the environment variables.  See the following on how to configure [this](https://docs.docker.com/install/linux/linux-postinstall/).
    ```bash
    sudo groupadd docker
    sudo usermod -aG docker $USER
    newgrp docker
    ```

    > Configure docker to run on boot
    ```bash
    sudo systemctl enable docker.service
    sudo systemctl enable containerd.service
    ```

2. Create your [Docker Hub account](https://hub.docker.com)

### Linux and MacOS

1. Go 1.15.8 [(instructions)](https://golang.org/doc/install#tarball).
   ```bash
   curl -L -O https://golang.org/dl/go1.15.8.linux-amd64.tar.gz
   sudo tar -xzf go1.15.8.linux-amd64.tar.gz 
   sudo chown -R root:root ./go
   sudo mv go /usr/local
   ```

   * Make sure that your GOPATH and PATH are configured correctly
   ```bash
   echo "export GOPATH=~/go" >> ~/.bashrc
   echo "export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin" >> ~/.bashrc
   source ~/.bashrc
   ```
   * You can verify it with the following commands:
   ```bash
   $ echo $GOPATH
   /home/YOURUSERNAME/go

   $ echo $PATH
   [...]:/home/YOURUSERNAME/go/bin:[...]
   ```

2. [Delve](https://github.com/go-delve/delve/tree/master/Documentation/installation) for Debugging
   ```bash
   go get github.com/go-delve/delve/cmd/dlv
   ```
3. Clone the repo and make it your working directory:
   ```bash
   git clone git@github.com:azure-octo/same-cli.git samecli
   cd samecli
   ```
4. Install Poetry for Python Package Management
   ```bash
   sudo apt install -y python3.8
   sudo apt install -y python3-pip
   curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python3 # requires Python3 is installed on your system
   echo "export PATH=$PATH:$HOME/.poetry/bin" >> ~/.bashrc
   source ~/.bashrc
   python3 -m pip install poetry
   python3 -m poetry install
   python3 -m poetry shell
   ```

4. Install pre-commit hooks
   ```bash
   pip install pre-commit
   pre-commit install
   go get 
   go get github.com/kisielk/errcheck
   go get honnef.co/go/tools/cmd/staticcheck
   go mod tidy
   go mod vendor
   ```
5. Install KIND # If you would like to do local development 
   From - https://kind.sigs.k8s.io/docs/user/quick-start/#installation
   ```bash
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.10.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind # Any other directory in your path will work.
   ```

6. Test out your repo
   ```bash
      cat "// your name <your_email>" >> AUTHORS.go
      git add AUTHORS.go
      git commit -m "Adding my name to the list of SAME authors"

   ```
7. Test your build # Will take a while the first time
   ```bash
      make all
      bin/same
   ```

You're ready to go!