# Local development

Start Jaeger locally to receive OpenTelemetry traces:

```bash
docker run -d --rm --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:1.76.0
```

Open [Jaeger UI](http://localhost:16686) after starting the service and sending a request.

Start Prometheus locally to scrape application metrics:

```bash
docker run -d --rm --name prometheus \
  -p 9090:9090 \
  -v "$(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  prom/prometheus
```

Open [Prometheus targets](http://localhost:9090/targets) to verify that the `infra-agent` target is up.

Start Loki locally to receive OpenTelemetry logs:

```bash
curl -L \
  https://raw.githubusercontent.com/grafana/loki/v3.7.0/cmd/loki/loki-local-config.yaml \
  -o loki-config.yaml

docker run -d --rm --name loki \
  -p 127.0.0.1:3100:3100 \
  -v "$(pwd)/loki-config.yaml:/mnt/config/loki-config.yaml:ro" \
  grafana/loki:3.7.0 \
  -config.file=/mnt/config/loki-config.yaml

curl http://localhost:3100/ready
```

The service sends OpenTelemetry logs directly to `http://localhost:3100/otlp/v1/logs`.

Start Grafana locally without persistent storage:

```bash
docker run -d --rm --name grafana \
  -p 127.0.0.1:3000:3000 \
  grafana/grafana-oss:latest
```

Open [Grafana](http://localhost:3000) and sign in with `admin` / `admin`. Configure the following data sources:

| Data source | URL |
| --- | --- |
| Loki | `http://host.docker.internal:3100` |
| Prometheus | `http://host.docker.internal:9090` |
| Jaeger | `http://host.docker.internal:16686` |

All Grafana users, data sources, and dashboards are deleted when the container stops.
