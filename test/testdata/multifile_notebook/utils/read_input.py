import kfp.components as components

def read_input(pipeline_params:dict,
               dataframe_path: components.OutputPath('CSV'),
               dataframe_long_path: components.OutputPath('CSV'),
               json_request_path: components.OutputPath('JSON'),
               parameters_path: components.OutputPath('JSON'),
               input_json_path: components.OutputPath('JSON'),
               model_json_path: components.OutputPath('JSON'),
               log_path: components.OutputPath('JSON'),
              ):
    
    data_folder = '/data/news_summarization'
    
    import os
    import json
    import pandas as pd
    import traceback
    import logging
    import time
    from datetime import datetime as dt
    
    from azure.storage.blob import BlobServiceClient
    
    script_start = time.time()
    
    logging.getLogger().setLevel(logging.INFO)
    
    print("Reading Input")
    
    #rest of code here