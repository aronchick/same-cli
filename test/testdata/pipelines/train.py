from kfp.components import OutputPath, InputPath

def train_model(train_data: InputPath(), mlpipeline_metrics_path: OutputPath('Metrics'),
          output_path: OutputPath(str), batch_size: int = 32, num_epochs: int = 250):

    # AUTOGENERATED! DO NOT EDIT! File to edit: 00_create_train.ipynb (unless otherwise specified).

    __all__ = ['train_dataset_url', 'train_dataset_fp', 'column_names', 'feature_names', 'label_name', 'class_names',
            'batch_size', 'train_dataset', 'pack_features_vector', 'train_dataset', 'model', 'predictions',
            'loss_object', 'loss', 'l', 'grad', 'optimizer', 'train_loss_results', 'train_accuracy_results',
            'num_epochs']

    import tensorflow as tf
    import json

    print("TensorFlow version: {}".format(tf.__version__))
    print("Eager execution: {}".format(tf.executing_eagerly()))

    train_dataset_fp = str(train_data)

    print("Local copy of the dataset file: {}".format(train_dataset_fp))

    # column order in CSV file
    column_names = ['sepal_length', 'sepal_width', 'petal_length', 'petal_width', 'species']

    feature_names = column_names[:-1]
    label_name = column_names[-1]

    print("Features: {}".format(feature_names))
    print("Label: {}".format(label_name))

    class_names = ['Iris setosa', 'Iris versicolor', 'Iris virginica']

    train_dataset = tf.data.experimental.make_csv_dataset(
        train_dataset_fp,
        batch_size,
        column_names=column_names,
        label_name=label_name,
        num_epochs=1)

    features, labels = next(iter(train_dataset))

    print(features)

    def pack_features_vector(features, labels):
        """Pack the features into a single array."""
        features = tf.stack(list(features.values()), axis=1)
        return features, labels

    train_dataset = train_dataset.map(pack_features_vector)

    features, labels = next(iter(train_dataset))

    print(features[:5])

    model = tf.keras.Sequential([
    tf.keras.layers.Dense(10, activation=tf.nn.relu, input_shape=(4,)),  # input shape required
    tf.keras.layers.Dense(10, activation=tf.nn.relu),
    tf.keras.layers.Dense(3)
    ])

    predictions = model(features)
    predictions[:5]

    tf.nn.softmax(predictions[:5])

    print("Prediction: {}".format(tf.argmax(predictions, axis=1)))
    print("    Labels: {}".format(labels))

    loss_object = tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True)

    def loss(model, x, y, training):
        # training=training is needed only if there are layers with different
        # behavior during training versus inference (e.g. Dropout).
        y_ = model(x, training=training)

        return loss_object(y_true=y, y_pred=y_)

    l = loss(model, features, labels, training=False)
    print("Loss test: {}".format(l))

    def grad(model, inputs, targets):
        with tf.GradientTape() as tape:
            loss_value = loss(model, inputs, targets, training=True)
        return loss_value, tape.gradient(loss_value, model.trainable_variables)

    optimizer = tf.keras.optimizers.SGD(learning_rate=0.01)

    loss_value, grads = grad(model, features, labels)

    print("Step: {}, Initial Loss: {}".format(optimizer.iterations.numpy(),
                                            loss_value.numpy()))

    optimizer.apply_gradients(zip(grads, model.trainable_variables))

    print("Step: {},         Loss: {}".format(optimizer.iterations.numpy(),
                                            loss(model, features, labels, training=True).numpy()))

    ## Note: Rerunning this cell uses the same model variables

    # Keep results for plotting
    train_loss_results = []
    train_accuracy_results = []

    for epoch in range(num_epochs):
        epoch_loss_avg = tf.keras.metrics.Mean()
        epoch_accuracy = tf.keras.metrics.SparseCategoricalAccuracy()

        # Training loop
        for x, y in train_dataset:
            # Optimize the model
            loss_value, grads = grad(model, x, y)
            optimizer.apply_gradients(zip(grads, model.trainable_variables))

            # Track progress
            epoch_loss_avg.update_state(loss_value)  # Add current batch loss
            # Compare predicted label to actual label
            # training=True is needed only if there are layers with different
            # behavior during training versus inference (e.g. Dropout).
            epoch_accuracy.update_state(y, model(x, training=True))

        # End epoch
        train_loss_results.append(epoch_loss_avg.result())
        train_accuracy_results.append(epoch_accuracy.result())

        if epoch % 50 == 0:
            print("Epoch {:03d}: Loss: {:.3f}, Accuracy: {:.3%}".format(epoch,
                                                                        epoch_loss_avg.result(),
                                                                        epoch_accuracy.result()))
    print("Epoch {:03d}: Loss: {:.3f}, Accuracy: {:.3%}".format(epoch,
                                                                        epoch_loss_avg.result(),
                                                                        epoch_accuracy.result()))
    metrics = {
        'metrics': [{
            'name': 'Loss',       # The name of the metric. Visualized as the column name in the runs table.
            'numberValue':  float(epoch_loss_avg.result()), # The value of the metric. Must be a numeric value.
            'format': "RAW",       # The optional format of the metric. Supported values are "RAW" (displayed in raw format) and "PERCENTAGE" (displayed in percentage format).
        },
        {
            'name': 'Accuracy',       # The name of the metric. Visualized as the column name in the runs table.
            'numberValue':  float(epoch_accuracy.result()), # The value of the metric. Must be a numeric value.
            'format': "PERCENTAGE",       # The optional format of the metric. Supported values are "RAW" (displayed in raw format) and "PERCENTAGE" (displayed in percentage format).
        }]
    }
    with open(mlpipeline_metrics_path, 'w') as f:
        json.dump(metrics, f)
    print("Saving model to {0}".format(output_path))
    model.save(output_path, include_optimizer=False, save_format='h5')