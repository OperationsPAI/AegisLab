#!/bin/bash

set -e

echo "clearing Redis database 0..."
docker exec redis redis-cli -n 0 FLUSHDB

echo "clearing all tables and views in MySQL rcabench database..."
docker exec mysql sh -c '
TABLES=$(mysql -uroot -pyourpassword -Nse "SELECT GROUP_CONCAT(table_name) FROM information_schema.tables WHERE table_schema=\"rcabench\" AND table_type=\"BASE TABLE\"")
VIEWS=$(mysql -uroot -pyourpassword -Nse "SELECT GROUP_CONCAT(table_name) FROM information_schema.tables WHERE table_schema=\"rcabench\" AND table_type=\"VIEW\"")
mysql -uroot -pyourpassword rcabench <<EOF
SET FOREIGN_KEY_CHECKS=0;
$([ -n "$TABLES" ] && echo "DROP TABLE IF EXISTS $TABLES;" || echo "")
$([ -n "$VIEWS" ] && echo "DROP VIEW IF EXISTS $VIEWS;" || echo "")
SET FOREIGN_KEY_CHECKS=1;
EOF
'
echo "Data cleanup completed."
