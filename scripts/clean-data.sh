#!/bin/bash

set -e

MYSQL_CONTAINER=${1:-mysql}
DB_NAME=${2:-rcabench}
DB_USER=${3:-root}
DB_PASS=${4:-yourpassword}

echo "clearing Redis database 0..."
docker exec redis redis-cli -n 0 FLUSHDB || echo "Redis not available or already empty"

echo "clearing etcd configs with prefix /rcabench/config/consumer/..."
docker exec etcd etcdctl del /rcabench/config/consumer/ --prefix || echo "etcd not available or no keys to delete"

echo "clearing all tables and views in MySQL ${DB_NAME} database..."
docker exec ${MYSQL_CONTAINER} sh -c '
TABLES=$(mysql -u'${DB_USER}' -p'${DB_PASS}' -Nse "SELECT GROUP_CONCAT(table_name) FROM information_schema.tables WHERE table_schema=\"'${DB_NAME}'\" AND table_type=\"BASE TABLE\"")
VIEWS=$(mysql -u'${DB_USER}' -p'${DB_PASS}' -Nse "SELECT GROUP_CONCAT(table_name) FROM information_schema.tables WHERE table_schema=\"'${DB_NAME}'\" AND table_type=\"VIEW\"")
mysql -u'${DB_USER}' -p'${DB_PASS}' '${DB_NAME}' <<EOF
SET FOREIGN_KEY_CHECKS=0;
$([ -n "$TABLES" ] && echo "DROP TABLE IF EXISTS $TABLES;" || echo "")
$([ -n "$VIEWS" ] && echo "DROP VIEW IF EXISTS $VIEWS;" || echo "")
SET FOREIGN_KEY_CHECKS=1;
EOF
' || echo "MySQL not available or no tables/views to drop"

echo "Data cleanup completed."
