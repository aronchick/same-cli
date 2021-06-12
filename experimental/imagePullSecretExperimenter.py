from __future__ import print_function
import time
import kubernetes.client
from pprint import pprint
from kubernetes import client, config
import base64
import json
import os
from dotenv import load_dotenv

config.load_kube_config()

# extras = connection.extra_dejson
# if extras.get("extra__kubernetes__in_cluster"):
#     config.load_incluster_config()
# elif extras.get("extra__kubernetes__kube_config") is None:
#     config.load_kube_config()
# else:
#     with tempfile.NamedTemporaryFile() as temp_config:
#         self.log.debug("loading kube_config from: connection kube_config")
#         temp_config.write(extras.get("extra__kubernetes__kube_config").encode())
#         temp_config.flush()
#         config.load_kube_config(temp_config.name)

# config.load_incluster_config()
v1 = client.CoreV1Api()
namespace = "kubeflow"
name = "SAME-EXPERIMENTNAME".lower()
metadata = {"name": name, "namespace": "kubeflow"}
api_version = "v1"
kind = "Secret"

load_dotenv()
pprint(os.environ)
docker_server = os.environ.get("SAME_REGISTRY_SERVER")

cred_payload = {
    "auths": {
        docker_server: {
            "username": os.environ.get("SAME_REGISTRY_READ_ONLY_SP_ID"),
            "password": os.environ.get("SAME_REGISTRY_READ_ONLY_SP_PASSWORD"),
            "email": os.environ.get("SAME_REGISTRY_READ_ONLY_EMAIL"),
            "auth": base64.b64encode(
                f'{os.environ.get("SAME_REGISTRY_READ_ONLY_SP_ID")}:{os.environ.get("SAME_REGISTRY_READ_ONLY_SP_PASSWORD")}'.encode()
            ).decode(),
        }
    }
}

data = {
    ".dockerconfigjson": base64.b64encode(json.dumps(cred_payload).encode()).decode()
}

secret = client.V1Secret(
    api_version="v1",
    data=data,
    kind="Secret",
    metadata=metadata,
    type="kubernetes.io/dockerconfigjson",
)
body = kubernetes.client.V1Secret(
    api_version, data, kind, metadata, type="kubernetes.io/dockerconfigjson"
)
api_response = None
try:
    pprint("Inside secret set")
    api_response = v1.create_namespaced_secret(namespace, body)
    # api_response = v1.read_namespaced_secret("pk-test-tls", namespace)
except kubernetes.client.rest.ApiException as e:
    pprint(f"Inside exception {e}")
    pprint(f"Status: {e.status}")
    if e.status == 409:
        pprint(f"Inside if {pprint(cred_payload)}")
        if (
            cred_payload["auths"]
            and cred_payload["auths"][docker_server]
            and cred_payload["auths"][docker_server]["username"]
            and cred_payload["auths"][docker_server]["password"]
            and cred_payload["auths"][docker_server]["email"]
        ):
            api_response = v1.replace_namespaced_secret(name, namespace, body)
        else:
            pprint(f"Missing value")
    else:
        raise e


# api_response = v1.create_namespaced_secret(namespace, body)

pprint(api_response)

"""
import numpy

kubectl create secret docker-registry gitlab-auth \
  --docker-server=$SAME_REGISTRY_SERVER \
  --docker-username=$SAME_REGISTRY_READ_ONLY_SP_ID \
  --docker-password=$SAME_REGISTRY_READ_ONLY_SP_PASSWORD \
  --docker-email=$SAME_REGISTRY_READ_ONLY_EMAIL \
  -n kubeflow
"""