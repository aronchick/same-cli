from kfp.components import OutputPath

def prepare_dataset(output_path: OutputPath(str)):

    import requests
    from pathlib import Path

    # Load raw data
    train_dataset_url = "https://storage.googleapis.com/download.tensorflow.org/data/iris_training.csv"
    print('Downloading dataset from {0} to {1}'.format(train_dataset_url, output_path))
    data = requests.get(train_dataset_url).content

    with open(output_path, 'wb') as writer:
        writer.write(data)
    
    print('Done.')

    # Modify raw data to select training data
    # ** not required **