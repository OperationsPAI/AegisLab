helm install ts1 train-ticket/trainticket -n ts1 --create-namespace \
  --set global.image.repository=docker.io/opspai \
  --set global.image.tag=v1.0.0-213-gf9294111 \
  --set services.tsUiDashboard.nodePort=30101 \
  --set global.security.allowInsecureImages=true

helm install ts0 train-ticket/trainticket -n ts0 --create-namespace \
  --set global.image.repository=pair-diagnose-cn-guangzhou.cr.volces.com/opspai \
  --set global.image.tag=v1.0.0-213-gf9294111 \
  --set services.tsUiDashboard.nodePort=30081 \
  --set mysql.image.repository=pair-diagnose-cn-guangzhou.cr.volces.com/library/mysql \
  --set rabbitmq.image.registry=pair-diagnose-cn-guangzhou.cr.volces.com \
  --set rabbitmq.image.repository=bitnamilegacy/rabbitmq \
  --set loadgenerator.image.repository=pair-diagnose-cn-guangzhou.cr.volces.com/opspai/loadgenerator \
  --set loadgenerator.initContainer.image=pair-diagnose-cn-guangzhou.cr.volces.com/nicolaka/netshoot:v0.14 \
  --set global.security.allowInsecureImages=true