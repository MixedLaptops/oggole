# Oggole Monitoring

Prometheus + Grafana monitoring for Oggole application.

## Setup

```bash
# On monitoring server
git clone https://github.com/MixedLaptops/oggole.git
cd oggole/monitoring
docker-compose up -d
```

## Access

- Prometheus: http://monitoring-server-ip:9090
- Grafana: http://monitoring-server-ip:3000 (admin/admin)

## Grafana Setup

1. Add Prometheus data source: `http://prometheus:9090`
2. Create dashboards for oggole metrics

## Stop

```bash
docker-compose down
```
