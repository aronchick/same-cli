def prepare_dataset(mnt_path: str = '/mnt/azure'):

    import requests
    from pathlib import Path

    base_path = Path(mnt_path).resolve(strict=False)
    print('Base Path:  {}'.format(base_path))
    data_path = base_path.joinpath('iris_training.csv').resolve(strict=False)

    # Load raw data
    train_dataset_url = "https://storage.googleapis.com/download.tensorflow.org/data/iris_training.csv"
    print('Downloading dataset from {0} to {1}'.format(train_dataset_url, str(data_path)))
    data = requests.get(train_dataset_url).content

    with open(str(data_path), 'wb') as writer:
        writer.write(data)
    
    print('Done.')

    # Modify raw data to select training data
    # ** not required **