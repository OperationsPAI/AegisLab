import subprocess
from loguru import logger
import os


ROOT_DIR = os.path.abspath(os.path.dirname(os.path.dirname(__file__)))


def build_image(dockerfile: str, tag: str, dir: str, build_args: dict = None):
    command = ["docker", "build", "-f", dockerfile, "-t", tag, dir]

    if build_args:
        for key, value in build_args.items():
            command.extend(["--build-arg", f"{key}={value}"])

    logger.info(f"Running command: {' '.join(command)}")

    result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    if result.returncode == 0:
        logger.info(f"Docker image {tag} built successfully")
    else:
        print(f"Failed to build image. Error: {result.stderr.decode('utf-8')}")

    return result
