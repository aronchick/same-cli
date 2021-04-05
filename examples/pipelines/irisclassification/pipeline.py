"""Main pipeline file"""
from typing import Dict
import kfp.dsl as dsl
import kfp.compiler as compiler
import kfp.components as components

@dsl.pipeline(
  name='Iris Classifaction',
  description='Iris Classifaction'
)
def irisclassification(
    epochs = 250,
    batch_size = 32,
):
  """Pipeline steps"""

  operations: Dict[str, dsl.ContainerOp] = dict()

  # preprocess data
  import data
  dataFactory = components.func_to_container_op(func=data.prepare_dataset,
  base_image='python:3.8-slim',
  packages_to_install=['pathlib2~=2.3.1', 'requests==2.25.0']
  )
  operations['preprocess'] = dataFactory()

  # train
  import train
  trainFactory = components.func_to_container_op(func=train.train_model,
  base_image='tensorflow/tensorflow:2.4.1',
  packages_to_install=['pathlib2>=2.3.1,<2.4.0', 'requests==2.25.0']
  )
  operations['training']= trainFactory(
      train_data=operations['preprocess'].output,
      batch_size=batch_size,
      num_epochs=epochs
  )
  operations['training'].after(operations['preprocess'])

if __name__ == '__main__':
  compiler.Compiler().compile(irisclassification, __file__ + '.tar.gz')