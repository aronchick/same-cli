import kfp
import kfp.dsl as dsl
from kfp.components import func_to_container_op, InputPath, OutputPath
import kfp.compiler as compiler
from kfp.dsl.types import Dict as KFPDict, List as KFPList

import function_0001
import function_0002
import function_0003
import function_0004


@dsl.pipeline(
    name="Compilation of pipelines",
)
def root():
    function_0001_op = func_to_container_op(
        func=function_0001.function_0001,
        base_image="python:3.9-slim-buster",
        packages_to_install=[
            "tensorflow",
            "plotly",
            # PIL", - this is detected as a package, but it's wrong  - should be "Pillow"
            "Pillow",
        ],
    )
    function_0001_task = function_0001_op()

    function_0002_op = func_to_container_op(
        func=function_0002.function_0002,
        base_image="python:3.9-slim-buster",
        packages_to_install=[
            "numpy",
            "matplotlib",
            "scipy",
            "requests",
            "pandas",
            # "io", - this was detected as a package, need to exclude
            # PIL", - this is detected as a package, but it's wrong  - should be "Pillow"
            "Pillow",
            "plotly",
            "chart_studio",
        ],
    )
    function_0002_task = function_0002_op()
    function_0002_task.after(function_0001_task)

    function_0003_op = func_to_container_op(
        func=function_0003.function_0003,
        base_image="python:3.9-slim-buster",
        packages_to_install=[],
    )
    function_0003_task = function_0003_op()
    function_0003_task.after(function_0002_task)

    function_0004_op = func_to_container_op(
        func=function_0004.function_0004,
        base_image="python:3.9-slim-buster",
        packages_to_install=[],
    )
    function_0004_task = function_0004_op()
    function_0004_task.after(function_0003_task)
