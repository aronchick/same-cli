"""Main pipeline file"""
from typing import Dict
import kfp.dsl as dsl
import kfp.compiler as compiler
import kfp.components as components
from kubernetes.client.models import V1EnvFromSource, V1SecretKeySelector

@dsl.pipeline(
  name='Iris Classifaction',
  description='Iris Classifaction'
)
def irisclassification(
    epochs = 250,
    batch_size = 32,
):
  """Pipeline steps"""

  pvcname = 'pipelinepvc'
  createStorageOp = dsl.VolumeOp(
  name = pvcname,
  resource_name = pvcname,
  size = '10Gi',
  storage_class = 'blob',
  modes = ['ReadWriteMany']
  )
  pipelinepvc = createStorageOp.volume

  storage_mount_path = '/mnt/azure'
  operations: Dict[str, dsl.ContainerOp] = dict()

  # preprocess data
  import data
  dataFactory = components.func_to_container_op(func=data.prepare_dataset,
  base_image='python:3.8-slim',
  packages_to_install=['pathlib2~=2.3.1', 'requests==2.25.0']
  )
  operations['preprocess'] = dataFactory(
      mnt_path=storage_mount_path
  )
  operations['preprocess'].after(createStorageOp)

  # train
  import train
  trainFactory = components.func_to_container_op(func=train.train_model,
  base_image='tensorflow/tensorflow:2.4.1',
  packages_to_install=['pathlib2>=2.3.1,<2.4.0', 'requests==2.25.0']
  )
  operations['training']= trainFactory(
      mnt_path=storage_mount_path,
      batch_size=batch_size,
      num_epochs=epochs
  )
  operations['training'].after(operations['preprocess'])

  for _, op in operations.items():
    op.add_pvolumes({storage_mount_path: pipelinepvc })

if __name__ == '__main__':
  compiler.Compiler().compile(irisclassification, __file__ + '.tar.gz')