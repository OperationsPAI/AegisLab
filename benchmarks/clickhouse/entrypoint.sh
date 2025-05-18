#!/bin/bash -ex
cd /app
echo "Running ClickHouse benchmark"
.venv/bin/python cli/prepare_inputs.py run
