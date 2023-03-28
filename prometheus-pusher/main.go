package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/numaproj/numaflow-go/pkg/sink/server"
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
)

type prometheusSink struct {
	logger     *zap.SugaredLogger
	skipFailed bool
	labels     map[string]string
	metrics    *MetricsPublisher
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

func (p *prometheusSink) push(msgPayloads []Payload) error {
	for _, payload := range msgPayloads {
		p.logger.Infof("Pushing : %v", payload)
		pusher, err := p.createPusher(fmt.Sprintf("%s_%s_%s", payload.Namespace, payload.Subsystem, payload.Name))
		if err != nil {
			return err
		}
		switch payload.Type {
		case "Gauge":
			p.logger.Debugf("Creating Collector %s", payload.Name)
			pusher = pusher.Collector(&myCollector{
				metric:     prometheus.NewDesc(payload.Name, "", nil, payload.Labels),
				metricType: prometheus.GaugeValue,
				value:      payload.Value,
			})
			appName := payload.Labels["app"]
			p.metrics.IncreaseAnomalyGenerated(payload.Namespace, appName, payload.Name)
		}
		err = pusher.Push()
		if err != nil {
			p.logger.Infof("Failed to push: %v error: %v", payload, err)
			return err
		}
		p.logger.Infof("Successfully pushed: %v", payload)
		p.metrics.IncreaseTotalSuccess()

	}
	return nil
}

func (p *prometheusSink) handle(ctx context.Context, datumStreamCh <-chan sinksdk.Datum) sinksdk.Responses {
	ok := sinksdk.ResponsesBuilder()
	failed := sinksdk.ResponsesBuilder()
	var payloads []string
	for datum := range datumStreamCh {
		payloads = append(payloads, string(datum.Value()))
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
		failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to push the metrics"))
	}
	var pls []Payload
	for _, payloadMsg := range payloads {
		p.metrics.IncreaseTotalPushed()
		var pl Payload
		err := json.Unmarshal([]byte(payloadMsg), &pl)
		if !p.skipFailed && err != nil {
			p.metrics.IncreaseTotalSkipped()
			return failed
		}
		pl.mergeLabels(p.labels)
		pls = append(pls, pl)
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
	var metricPort int
	meticslabels := numaflag.MapFlag{}

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

	ps := prometheusSink{logger: logger, skipFailed: skipFailed, labels: labels}
	ps.metrics = NewMetricsServer(labels)
	go ps.metrics.startMetricServer(metricPort)
	ps.logger.Infof("Metrics publisher initialized with port=%d", metricPort)

	server.New().RegisterSinker(sinksdk.SinkFunc(ps.handle)).Start(context.Background())
}
