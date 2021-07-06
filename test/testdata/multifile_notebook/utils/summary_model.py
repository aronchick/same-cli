import kfp.components as components

def summary_model(dataframe_path: components.InputPath('CSV'),
                  model_json_path: components.InputPath('JSON'),
                  log_path: components.InputPath('JSON'),
                  summary_list_path: components.OutputPath('JSON')
                 ):
    import os
    import re
    import json
    import time
    import numpy as np
    import pandas as pd
        
    import spacy
    import kneed
        
    # logging specific packages
    import traceback
    from datetime import datetime as dt
    import logging
    logging.getLogger().setLevel(logging.INFO)
        

    print("Summary Model")
    
    #rest of code here