import kfp.components as components

def zip_files(pipeline_params:dict,
              dataframe_path: components.InputPath('CSV'),
              dataframe_long_path: components.InputPath('CSV'),
              json_request_path: components.InputPath('JSON'),
              parameters_path: components.InputPath('JSON'),
              input_json_path: components.InputPath('JSON'),
              model_json_path: components.InputPath('JSON'),
              log_path: components.InputPath('JSON'),
              score_list_path: components.InputPath('JSON'),
              summary_list_path: components.InputPath('JSON'),
              steep_list_path: components.InputPath('JSON'),
              ner_df_path: components.InputPath('CSV'),
              zip_output_path: components.OutputPath('JSON')
             ):
    import os
    import uuid
    import json
    import time
    import base64
    import shutil
    import logging
    import zipfile
    import numpy as np
    import pandas as pd
    import requests
    import traceback
    from datetime import datetime as dt
    
    logging.getLogger().setLevel(logging.INFO)
    
    print("Zipping Files")
    
    #rest of code here