#!/bin/bash

set -e

NAMESPACE=${1:-exp}
MYSQL_POD=${2:-rcabench-mysql-0}
DB_NAME=${3:-rcabench}
DB_USER=${4:-root}
DB_PASS=${5:-yourpassword}

echo "Clearing Redis database 0..."
kubectl exec -it rcabench-redis-0 -n ${1:-exp} -- redis-cli -n 0 FLUSHDB || echo "Redis not available or already empty"

echo "Clearing etcd configs..."
kubectl exec -it rcabench-etcd-0 -n ${NAMESPACE} -- etcdctl del /rcabench/config/producer/ --prefix || echo "etcd not available or keys with prefix /rcabench/config/producer/ do not exist"
kubectl exec -it rcabench-etcd-0 -n ${NAMESPACE} -- etcdctl del /rcabench/config/consumer/ --prefix || echo "etcd not available or keys with prefix /rcabench/config/consumer/ do not exist"
kubectl exec -it rcabench-etcd-0 -n ${NAMESPACE} -- etcdctl del /initialized_consumer || echo "etcd not available or /initialized_consumer key not present"
kubectl exec -it rcabench-etcd-0 -n ${NAMESPACE} -- etcdctl del /initialized_producer || echo "etcd not available or /initialized_producer key not present"

echo "Clearing all tables and views in MySQL ${DB_NAME} database..."
kubectl exec -it ${MYSQL_POD} -n ${NAMESPACE} -- sh -c '
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
