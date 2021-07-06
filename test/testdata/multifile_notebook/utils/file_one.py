from pandas import DataFrame

from mimesis import Person
from mimesis import Address
from mimesis.enums import Gender
from mimesis import Datetime
from mimesis import Text
import pandas as pd
import random
import sys


def create_rows_mimesis(num=1):
    person = Person("en")
    addess = Address()
    datetime = Datetime()
    text = Text()

    output = [
        {
            "name": person.full_name(gender=Gender.FEMALE),
            "address": addess.address(),
            "name": person.name(),
            "email": person.email(),
            "city": addess.city(),
            "state": addess.state(),
            "date_time": datetime.datetime(),
            "randomdata_range": random.randint(1000, 2000),
            "randomdata_float": random.random(),
            "randomdata_text": text.text(20),
        }
        for x in range(num)
    ]
    return output


def function_one(number_of_examples) -> DataFrame:
    import numpy as np
    import pandas as pd

    print("Function One")
    df = pd.DataFrame(create_rows_mimesis(number_of_examples))
    print(f"DF Head: {df.head()}")
    # rest of code here

    print(f"DF Size: {len(df.index)}")
    return df
