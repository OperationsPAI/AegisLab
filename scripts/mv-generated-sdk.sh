#!/bin/bash -ex
DST="sdk/python/src/rcabench/openapi"
rm -rf $DST
rm -rf sdk/python/src/rcabench/docs
cp -r sdk/python-gen/openapi $DST
cp -r sdk/python-gen/docs sdk/python/src/rcabench

find $DST -name "*.py" -type f -exec sed -i 's/openapi/rcabench.openapi/g' {} \;
find sdk/python/src/rcabench -name "*.md" -type f -exec sed -i 's/openapi/rcabench.openapi/g' {} \;

rm $DST/py.typed