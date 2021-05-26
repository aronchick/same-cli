# How to run SAME against AML

- Create an Azure ML Workspace
- Create an Azure Service Principal - https://github.com/Azure/MachineLearningNotebooks/blob/master/how-to-use-azureml/manage-azureml-service/authentication-in-azureml/authentication-in-azureml.ipynb
- Create a JSON file (creds.json) of the following format - make sure to add this file to your .gitignore:
{
"AML_SP_NAME":"<SERVICE_PRINCIPAL_NAME>",
AML_SP_APP_ID:"<SERVICE_PRINCIPAL_APP_ID>",
AML_SP_OBJECT_ID:"<SERVICE_PRINCIPAL_OBJECT_ID>",
AML_SP_TENANT_ID:"<SERVICE_PRINCIPAL_TENANT_ID>",
AML_SP_PASSWORD_VALUE:"<SERVICE_PRINCIPAL_PASSWORD>",
AML_SP_PASSWORD_ID:"<SERVICE_PRINCIPAL_PASSWORD_ID>"
}

- Put this in aa variable with the following command:
`export AML_SP_CREDENTIALS=``cat creds.json```
- Run your pipeline with the following command:

`same program run --target=aml --run-params=$AML_SP_CREDENTIALS`
