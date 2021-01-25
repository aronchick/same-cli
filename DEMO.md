Show initial getting started:
- `same init` a new directory
- open a jupyter notebook
- add some code
- check it in (executes some GHA)

Change to new example:
- `same create program -f same.yaml` - deploys:
  - AKS
  - Disk
  - Kubeflow to AKS
  - PV against the disk
  - pipeline (pre-compiled) to Kubeflow that knows how to use that disk
  - NTH: Copies public CSV file to disk
- Open the dashboard, show the pipeline and parameters
- Back to the command line, `same run program foobaz_pipeline_name --params=foo` => executes on Kubeflow
  - NTH: checks model into GitHub? check into AML?
  - NTH: show metadata
- Back to `same.yaml` - switch to GCP - redeploy
- Show that everything came back up - run the experiment again
- Now change the parameters - `same run program foobaz_pipeline_name --params=foo`

Nice to have:
- Do the above against KIND