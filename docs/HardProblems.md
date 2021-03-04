The path to SAME will be challenged in a series of different domains. Initial problems we see include:

1. How do we run code in a cluster such that a data scientist doesn't need to know about containerization? How do they fork public repo and change code in as small a way as possible? How do we check your cluster is appropriate? Etc
The current experience for a data scientist includes:
- Setting up a local jupyter notebook (`pip install jupyter; jupyter notebook`)
- Opening a browser to make edits to the notebook
- Mounting/installing new libraries via that notebook (``)

2. What do we do about pipelines? Kubeflow requires compilation which isn't great. Kedro does it all in python which feels right, how do we test/swap?
3. Data versioning - how do we capture the data this ran on? How do we capture and verify data schema? Give errors when schema invalid? How do we trivially allow swapping of data sources for trying existing models in new organizations? (including access to the data)
