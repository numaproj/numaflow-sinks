# Prometheus Pusher
This Sink will send the prometheus metrics `push gateway`

## Environment Variables

	PROMETHEUS_SERVER      : Prometheus or Push Gateway URL
	SKIP_VALIDATION_FAILED : Skip the marshal error for prometheus metric
    METRICS_LABELS         : Configure additional Labels for metrics
    EXCLUDE_LABELS         : Remove the Labels from Metrics

### Example Configuration

```yaml
 - name: prometheus-pusher
    sink:
      udsink:
        container:
          env:
          - name: SKIP_VALIDATION_FAILED
            value: "true"
          - name: "PROMETHEUS_SERVER"
            value: "pushgateway.monitoring.svc.cluster.local:9091"
          - name: "METRICS_LABELS"
            value: "label1=value1,label2=value2"
          - name: "EXCLUDE_LABELS"
            value: "label3,label4,label5"
          image: quay.io/numaio/numaflow-sink/prometheus-pusher:latest

```