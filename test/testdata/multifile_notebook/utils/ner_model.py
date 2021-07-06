import kfp.components as components

def ner_model(dataframe_path: components.InputPath('CSV'),
              model_json_path: components.InputPath('JSON'),
              log_path: components.InputPath('JSON'),
              ner_df_path: components.OutputPath('CSV')
             ):
    import os
    import json
    import time
    import numpy as np
    import pandas as pd
    
    import spacy
    
    # logging specific packages
    import traceback
    from datetime import datetime as dt
    import logging
    logging.getLogger().setLevel(logging.INFO)
    
    print("Ner Model")
    
    #rest of code here