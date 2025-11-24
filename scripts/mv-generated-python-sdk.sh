#!/bin/bash -ex
DST="sdk/python/src/rcabench/openapi"
rm -rf $DST
rm -rf sdk/python/docs
rm -f sdk/python/pyproject.toml
cp -r sdk/python-gen/openapi $DST
cp -r sdk/python-gen/docs sdk/python
cp -r sdk/python-gen/pyproject.toml sdk/python

find $DST -name "*.py" -type f -exec sed -i 's/openapi/rcabench.openapi/g' {} \;
find sdk/python/src/rcabench -name "*.md" -type f -exec sed -i 's/openapi/rcabench.openapi/g' {} \;

rm $DST/py.typed