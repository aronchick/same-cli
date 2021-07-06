import kfp.components as components

def steep_model(dataframe_path: components.InputPath('CSV'),
                model_json_path: components.InputPath('JSON'),
                log_path: components.InputPath('JSON'),
                steep_list_path: components.OutputPath('JSON')
               ):
    import os
    import re
    import json
    import time
    import pickle
    import numpy as np
    import pandas as pd
    import sklearn
    import traceback
    from datetime import datetime as dt
    
    import logging
    logging.getLogger().setLevel(logging.INFO)
    
    print("Steep Model")
    
    #rest of code here