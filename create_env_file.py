import os
from pathlib import Path

variables_to_output = {}
for key in os.environ:
    if key.startswith("AZURE_") or key.startswith("SAME_"):
        variables_to_output[key] = os.environ.get(key)

Path('.env').unlink()

with Path('.env').open('w') as f:
    for key in variables_to_output:
        f.write(f'{key}="{variables_to_output[key]}"\n')

