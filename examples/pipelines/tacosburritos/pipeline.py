"""Main pipeline file"""
import kfp.dsl as dsl
import kfp.compiler as compiler
import kfp.components as components

from kubernetes.client.models.v1_volume import V1Volume

@dsl.pipeline(
  name='Tacos vs. Burritos',
  description='Simple TF CNN'
)
def tacosandburritos_train(
    epochs = 5,
    batch = 32,
    learning_rate = 0.0001
):
  """Pipeline steps"""

  persistent_volume_path = '/mnt/azure'
  data_download = 'https://aiadvocate.blob.core.windows.net/public/tacodata.zip'
  model_name = 'tacosandburritos'
  profile_name = 'tacoprofile'
  operations = {}
  image_size = 160
  training_folder = 'train'
  training_dataset = 'train.txt'
  model_folder = 'model'

  pvcname = 'pipelinepvc'
  createStorageOp = dsl.VolumeOp(
  name = pvcname,
  resource_name = pvcname,
  size = '10Gi',
  storage_class = 'blob',
  modes = ['ReadWriteMany']
  )
  pipelinepvc = createStorageOp.volume

  # preprocess data
  import data
  dataFactory = components.func_to_container_op(func=data.executeData,
  base_image='tensorflow/tensorflow:2.0.0a0-py3',
  packages_to_install=['pathlib2', 'requests', 'wget']
  )
  operations['preprocess'] = dataFactory(
      base_path=persistent_volume_path,
      data=training_folder,
      target=training_dataset,
      img_size=image_size,
      zipfile=data_download
  )
  operations['preprocess'].after(createStorageOp)

  # train
  import train
  trainFactory = components.func_to_container_op(func=train.executeTrain,
  base_image='tensorflow/tensorflow:2.0.0a0-py3',
  packages_to_install=['Pillow', 'pathlib2']
  )
  operations['training']= trainFactory(
      base_path=persistent_volume_path,
      data=training_folder,
      epochs=epochs,
      batch=batch,
      image_size=image_size,
      lr=learning_rate,
      outputs=model_folder,
      dataset=training_dataset
  )
  operations['training']
  operations['training'].after(operations['preprocess'])

  for _, op in operations.items():
    # op.container.set_image_pull_policy("Always")
    op.add_pvolumes({persistent_volume_path: pipelinepvc })

if __name__ == '__main__':
  compiler.Compiler().compile(tacosandburritos_train, __file__ + '.tar.gz')