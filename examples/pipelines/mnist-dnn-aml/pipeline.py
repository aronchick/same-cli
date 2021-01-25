import os

import kfp
import kfp.dsl as dsl

from kubernetes.client.models import V1Volume, V1VolumeMount, V1PersistentVolumeClaimSource


@kfp.dsl.pipeline(
    name='Test',
    description='Test'
)
def interestingpipe(someparam='test'):

    # TODO: Verify this works.
    vol = kfp.dsl.VolumeOp(
      resource_name = "pipelinepvc",
      size = "2Gi",
      # no storageclass, for dynamic creation only
      modes = "ReadWriteMany",
      volume_name = "THE_VOLUME_CREATED_BY_TERRAFORM"
    )

    dataprepfunc = kfp.components.func_to_container_op(.....)

    dataprep = dataprepfunc("somearg")
    dataprep.add_volume(V1Volume(persistent_voume_claim=V1PersisentVolumeClaimSource(claim_name="pipelinepvc")))
    dataprep.add_volume_mount(V1VolumeMount(mount_path='/data', name="pipelinepvc"))

    # Loads the distributed tensorflow component
    # TODO: Check whether it needs to be the raw URL instead
    tfjobcomponent = kfp.components.load_component_from_url("https://raw.githubusercontent.com/kubeflow/pipelines/1.2.0/components/kubeflow/dnntrainer/component.yaml")
    trainer = tfjobcomponent()
    trainer.add_volume(V1Volume(persistent_voume_claim=V1PersisentVolumeClaimSource(claim_name="pipelinepvc")))
    trainer.add_volume_mount(V1VolumeMount(mount_path='/data', name="pipelinepvc"))


    for _, op in operations.items():
        op.container.set_image_pull_policy("Always")


if __name__ == '__main__':
   import kfp.compiler as compiler
   compiler.Compiler().compile(iris_classifier_train, __file__ + '.tar.gz')
