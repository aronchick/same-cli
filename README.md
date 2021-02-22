# PLEASE ONLY CONTRIBUTE WITH THE EXPECTATION THAT 100% OF THIS REPO TO BE DELETED BY 2021-05-01.
- We're big on keeping our commitments, but we may break this one
- that said, it's a good assumption that nothing you do here will last in any way other than for archival purposes

# TO GET STARTED
Read up on our development environment (here)[docs/development/satup-same-development-env.md] 

# User experience
Our goal is to have a user experience that looks like this:
```
# Local experience
# Install same CLI & configure K8s cluster (as easy as pip install foobaz)
$ export SAME_ENVIRONMENT=local
$ curl http://projectsame.org/install.sh | bash
```
The above:
- Downloads and installs a CLI tool
- Installs a local kubernetes API (e.g., via k3s/k3d)
- Installs Kubeflow into that cluster
- Provides an SAME endpoint inside the Kubernetes cluster that can respond to `same` commands

```
# Create a program in Kubernetes and populates it with the defaults generated
# by the original author. Also uploads history exported by the original author
$ same create program -f https://github.com/uoftoronto/papers/arXiv.5778v1:1.0
```
The above:
- Creates all necessary services inside the Kubernetes cluster (same experience on local and/or hosted)
- Provisions resources necessary (e.g. Premium disk, GPU notes, etc) OR clearly alerts the user that the expected provisioning is not possible
- Pushes all metadata into the chosen metadata store for future comparison 

```
# Execute a pipeline with new parameters defined by the original author
$ same run program arXiv.1303.5778v1:1.0 --epochs=1000
```
The above: 
- Executes the program previously uploaded with the override parameter of 'epochs=100'
- Remainder of the execution occurs as expectedwith the defaults
- Is _usually_ driven off a git like system
- Pushes results into expected metadata store

```
# Export the entire pipeline and history to a single file. Results in file 
# arXiv.1303.5778v1:1.1.tgz.
$ same export program -n arXiv.1303.5778v1:1.1
```
The above:
- Includes all necessary services and infrastructure descriptions to recreate the system anywhere
- Also offers a way to select metadata to export and include
- Does not NECESSARILY include data, but could include pointers to the data


With all this, we think we can make a breakthrough in the way that machine learning is reproduced across multiple environments.


=====
ICEBERGS:
1. How do we run code in a cluster such that a data scientist doesn't need to know about containerization? How do they fork public repo and change code in as small a way as possible? How do we check your cluster is appropriate? Etc
2. What do we do about pipelines? Kubeflow requires compilation which isn't great. Kedro does it all in python which feels right, how do we test/swap?
3. Data versioning - how do we capture the data this ran on? How do we capture and verify data schema? Give errors when schema invalid? How do we trivially allow swapping of data sources for trying existing models in new organizations? (including access to the data)