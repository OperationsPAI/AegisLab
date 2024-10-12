from experiments.utils import build_image
import os
from loguru import logger


def build_benchmark_images(algo_name):
    builder_image = f"{algo_name}_builder:local"
    data_builder_image = f"{algo_name}_data_builder:local"
    runner_image = f"{algo_name}_runner:local"
    executor_image = f"{algo_name}_executor:local"
    logger.info("Starting the Docker build process...")

    # Build the builder image
    logger.info("Building the builder image...")
    builder_dockerfile = os.path.join("algorithms", algo_name, "builder.Dockerfile")
    builder_dir = os.path.join("algorithms", algo_name)
    build_image(dockerfile=builder_dockerfile, tag=builder_image, dir=builder_dir)

    # Build the data builder image
    logger.info("Building the data builder image...")
    data_builder_dockerfile = os.path.join("benchmarks", "clickhouse", "Dockerfile")
    data_builder_dir = os.path.join("benchmarks", "clickhouse")
    build_image(
        dockerfile=data_builder_dockerfile, tag=data_builder_image, dir=data_builder_dir
    )

    # Build the runner image with build arguments
    logger.info("Building the runner image...")
    runner_dockerfile = os.path.join("algorithms", algo_name, "runner.Dockerfile")
    runner_dir = os.path.join("algorithms", algo_name)
    build_args = {
        "BUILDER_IMAGE": builder_image,
        "DATA_BUILDER_IMAGE": data_builder_image,
    }
    build_image(
        dockerfile=runner_dockerfile,
        tag=runner_image,
        dir=runner_dir,
        build_args=build_args,
    )

    # Build the execution image with build arguments
    logger.info("Building the final execution image...")
    executor_dockerfile = os.path.join("experiments", "execution.Dockerfile")
    build_args = {
        "RUNNER_IMAGE": runner_image,
    }
    build_image(
        dockerfile=executor_dockerfile,
        tag=executor_image,
        dir="experiments",
        build_args=build_args,
    )

    logger.info("All Docker images built successfully.")


build_benchmark_images("e-diagnose")
