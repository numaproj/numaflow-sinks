package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"
	numaflag "github.com/numaproj/numaflow-sinks/shared/flag"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/zap"
)

const (
	PROMETHEUS_SERVER      = "PROMETHEUS_SERVER"
	SKIP_VALIDATION_FAILED = "SKIP_VALIDATION_FAILED"
	METRICS_LABELS         = "METRICS_LABELS"
	METRICS_NAME           = "METRICS_NAME"
	EXCLUDE_METRIC_LABELS  = "EXCLUDE_METRICS_LABELS"
)

type prometheusSink struct {
	logger               *zap.SugaredLogger
	skipFailed           bool
	labels               map[string]string
	excludeLabels        []string
	metrics              *MetricsPublisher
	ignoreMetricsTs      bool
	metricsName          string
	enableMsgTransformer bool
}

type myCollector struct {
	metric     *prometheus.Desc
	ts         time.Time
	metricType prometheus.ValueType
	value      float64
}

func (c *myCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metric
}

func (c *myCollector) Collect(ch chan<- prometheus.Metric) {
	var metric prometheus.Metric
	if c.ts.IsZero() {
		metric = prometheus.MustNewConstMetric(c.metric, c.metricType, c.value)
	} else {
		metric = prometheus.NewMetricWithTimestamp(c.ts, prometheus.MustNewConstMetric(c.metric, c.metricType, c.value))
	}
	ch <- metric
}

func (p *prometheusSink) push(msgPayloads []PrometheusPayload) error {
	for _, payload := range msgPayloads {
		p.logger.Debugw("Pushing PrometheusPayload ", zap.Any("payload", payload))
		pusher, err := p.createPusher(fmt.Sprintf("%s_%s_%s", payload.Namespace, payload.Subsystem, payload.Name))
		if err != nil {
			return err
		}
		switch payload.Type {
		case "Gauge":
			p.logger.Debugf("Creating Collector %s", payload.Name)
			if p.ignoreMetricsTs {
				pusher = pusher.Collector(&myCollector{
					metric:     prometheus.NewDesc(payload.Name, "", nil, nil),
					metricType: prometheus.GaugeValue,
					value:      payload.Value,
				})
			} else {
				pusher = pusher.Collector(&myCollector{
					metric:     prometheus.NewDesc(payload.Name, "", nil, nil),
					metricType: prometheus.GaugeValue,
					value:      payload.Value,
					ts:         time.UnixMilli(payload.TimestampMs),
				})
			}

			for key, value := range payload.Labels {
				pusher.Grouping(key, value)
			}
			appName := payload.Labels["app"]
			p.metrics.IncreaseAnomalyGenerated(payload.Namespace, appName, payload.Name)
		default:
			p.logger.Errorw("Unsupported Metrics Type", zap.Any("payload", payload))
			return fmt.Errorf("unsupported Metrics Type")
		}
		err = pusher.Push()
		if err != nil {
			p.logger.Errorw("Failed to push", zap.Any("payload", payload), zap.Error(err))
			return err
		}
		p.logger.Infow("Successfully pushed", zap.Any("payload", payload))
		p.metrics.IncreaseTotalSuccess()

	}
	return nil
}

func (p *prometheusSink) Sink(ctx context.Context, datumStreamCh <-chan sinksdk.Datum) sinksdk.Responses {
	ok := sinksdk.ResponsesBuilder()
	failed := sinksdk.ResponsesBuilder()
	var payloads []string
	for datum := range datumStreamCh {
		payloads = append(payloads, string(datum.Value()))
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
		failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to push the metrics"))
	}
	var pls []PrometheusPayload
	var prometheusPayload PrometheusPayload
	for _, payloadMsg := range payloads {
		p.metrics.IncreaseTotalPushed()
		if p.enableMsgTransformer {
			var opl OriginalPayload
			err := json.Unmarshal([]byte(payloadMsg), &opl)
			if !p.skipFailed && err != nil {
				p.metrics.IncreaseTotalSkipped()
				return failed
			}
			prometheusPayload = *opl.ConvertToPrometheusPayload(p.metricsName)
		} else {
			err := json.Unmarshal([]byte(payloadMsg), &prometheusPayload)
			if !p.skipFailed && err != nil {
				p.metrics.IncreaseTotalSkipped()
				return failed
			}
		}
		prometheusPayload.mergeLabels(p.labels)

		if len(p.excludeLabels) > 0 {
			prometheusPayload.excludeLabels(p.excludeLabels)
		}

		pls = append(pls, prometheusPayload)
	}
	err := p.push(pls)
	if err != nil {
		p.metrics.IncreaseTotalFailed()
		p.logger.Errorf("Failed to push the Metrics", zap.Error(err))
		return failed
	}
	p.metrics.IncreaseTotalSuccess()
	return ok
}

func (p *prometheusSink) createPusher(jobName string) (*push.Pusher, error) {
	server, ok := os.LookupEnv(PROMETHEUS_SERVER)
	if !ok {
		return nil, fmt.Errorf("Prometheus URL not found")
	}
	pusher := push.New(server, jobName)
	return pusher, nil
}

func parseStringToSlice(envValue string) []string {
	return strings.Split(envValue, ",")
}

func parseStringToMap(envValue string) map[string]string {
	items := make(map[string]string)
	if envValue == "" {
		return items
	}
	datas := strings.Split(envValue, ",")
	getKeyVal := func(item string) (key, val string) {
		splits := strings.Split(item, "=")
		if len(splits) == 2 {
			key = splits[0]
			val = splits[1]
		}
		return
	}
	for _, item := range datas {
		key, val := getKeyVal(item)
		if key != "" {
			items[key] = val
		}
	}
	return items
}

func main() {
	logger := logging.NewLogger().Named("prometheus-sink")
	skipFailedStr := os.Getenv(SKIP_VALIDATION_FAILED)
	labels := parseStringToMap(os.Getenv(METRICS_LABELS))
	excludeLabels := parseStringToSlice(os.Getenv(EXCLUDE_METRIC_LABELS))
	metricName := os.Getenv(METRICS_NAME)
	if metricName == "" {
		metricName = "namespace_app_rollouts_unified_anomaly"
	}
	var metricPort int
	var ignoreMetricsTs, enableMsgTransformer bool
	meticslabels := numaflag.MapFlag{}

	flag.BoolVar(&enableMsgTransformer, "enableMsgTransformer", false, "Enable Prometheus message Transformer")
	flag.BoolVar(&ignoreMetricsTs, "ignoreMetricsTs", true, "Ignore Metrics Timestamp")
	flag.IntVar(&metricPort, "udsinkMetricsPort", 9090, "Metrics Port")
	flag.Var(&meticslabels, "udsinkMetricsLabels", "Sink Metrics Labels E.g: label=val1,label1=val2")
	// Parse the flag
	flag.Parse()
	skipFailed := false
	var err error
	if skipFailedStr != "" {
		skipFailed, err = strconv.ParseBool(skipFailedStr)
		if err != nil {
			panic(err)
		}
	}

	ps := prometheusSink{logger: logger, skipFailed: skipFailed, labels: labels, excludeLabels: excludeLabels, ignoreMetricsTs: ignoreMetricsTs,
		metricsName: metricName, enableMsgTransformer: enableMsgTransformer}
	ps.metrics = NewMetricsServer(labels)
	go ps.metrics.startMetricServer(metricPort)
	ps.logger.Infof("Metrics publisher initialized with port=%d", metricPort)
	err = sinksdk.NewServer(&ps).Start(context.Background())
	if err != nil {
		log.Panic("Failed to start sink function server: ", err)
	}
}
