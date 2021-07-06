import kfp.components as components
    
def sentiment_model(dataframe_long_path: components.InputPath('CSV'),
                    input_json_path: components.InputPath('JSON'),
                    log_path: components.InputPath('JSON'),
                    score_list_path: components.OutputPath('JSON')
                    ):
    import os
    import json
    import time
    import numpy as np
    import pandas as pd
    
    import tensorflow as tf
    import keras
    from keras.models import model_from_json
    from keras.preprocessing.sequence import pad_sequences

    # logging specific packages
    from datetime import datetime as dt
    import traceback
    import logging
    logging.getLogger().setLevel(logging.INFO)

    #####section: not sure how to deal with this os here.

    print("Sentiment Model")
    
    #rest of code here