def executeData(base_path: str = '../../data', data: str = 'train', target: str ='train.txt',
            img_size: int = 160, zipfile: str = '../../tacodata.zip',
            force: bool = False):

    import os
    import shutil
    import zipfile
    import argparse
    from pathlib2 import Path
    import wget
    import tensorflow as tf


    def check_dir(path):
        if not os.path.exists(path):
            os.makedirs(path)
        return Path(path).resolve(strict=False)


    def download(source, target, force_clear=False):
        if force_clear and os.path.exists(target):
            print('Removing {}...'.format(target))
            shutil.rmtree(target)

        check_dir(target)

        targt_file = str(Path(target).joinpath('data.zip'))
        if os.path.exists(targt_file) and not force_clear:
            print('data already exists, skipping download')
            return

        if source.startswith('http'):
            print("Downloading from {} to {}".format(source, target))
            wget.download(source, targt_file)
            print("Done!")
        else:
            print("Copying from {} to {}".format(source, target))
            shutil.copyfile(source, targt_file)

        print('Unzipping {}'.format(targt_file))
        zipr = zipfile.ZipFile(targt_file)
        zipr.extractall(target)
        zipr.close()


    def process_image(path, image_size=160):
        img_raw = tf.io.read_file(path)
        img_tensor = tf.image.decode_jpeg(img_raw, channels=3)
        img_final = tf.image.resize(img_tensor, [image_size, image_size]) / 255
        return img_final


    def walk_images(path, image_size=160):
        imgs = []
        print('Scanning {}'.format(path))
        # find subdirectories in base path
        # (they should be the labels)
        labels = []
        for (_, dirs, _) in os.walk(path):
            print('Found {}'.format(dirs))
            labels = dirs
            break

        for d in labels:
            tmp_path = os.path.join(path, d)
            print('Processing {}'.format(tmp_path))
            # only care about files in directory
            for item in os.listdir(tmp_path):
                if not item.lower().endswith('.jpg'):
                    print('skipping {}'.format(item))
                    continue

                image = os.path.join(tmp_path, item)
                try:
                    img = process_image(image, image_size)
                    assert img.shape[2] == 3, "Invalid channel count"
                    # write out good images
                    imgs.append(image)
                except img.shape[2] != 3:
                    print('{}\n'.format(image))

        return imgs


    print('Using TensorFlow v.{}'.format(tf.__version__))

    base_path = Path(base_path).resolve(strict=False)
    print('Base Path:  {}'.format(base_path))
    data_path = base_path.joinpath(data).resolve(strict=False)
    print('Train Path: {}'.format(data_path))
    target_path = Path(base_path).resolve(strict=False).joinpath(target)
    print('Train File: {}'.format(target_path))

    print('Acquiring data...')
    download('https://aiadvocate.blob.core.windows.net/public/tacodata.zip',
             str(base_path), force)

    if os.path.exists(str(target_path)):
        print('dataset text file already exists, skipping check')
    else:
        print('Testing images...')
        images = walk_images(str(data_path), img_size)

        # save file
        print('writing dataset to {}'.format(target_path))
        with open(str(target_path), 'w+') as f:
            f.write('\n'.join(images))