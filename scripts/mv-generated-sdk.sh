#!/bin/bash -ex
DST="sdk/python/src/rcabench/openapi"
rm -rf $DST
cp -r sdk/python-gen/rcabench_client $DST
find $DST -name "*.py" -type f -exec sed -i 's/rcabench_client/rcabench.openapi/g' {} \;
rm $DST/py.typed
