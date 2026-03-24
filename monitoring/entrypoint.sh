#!/bin/sh
set -e

cp /etc/prometheus/prometheus.template.yml /etc/prometheus/prometheus.yml
sed -i "s#GRAFANA_REMOTE_WRITE_URL_PLACEHOLDER#$GRAFANA_REMOTE_WRITE_URL#g" /etc/prometheus/prometheus.yml
sed -i "s#GRAFANA_USER_ID_PLACEHOLDER#$GRAFANA_USER_ID#g" /etc/prometheus/prometheus.yml
sed -i "s#GRAFANA_API_KEY_PLACEHOLDER#$GRAFANA_API_KEY#g" /etc/prometheus/prometheus.yml
sed -i "s#ARION_URL_PLACEHOLDER#$ARION_URL#g" /etc/prometheus/prometheus.yml

exec prometheus --config.file=/etc/prometheus/prometheus.yml
