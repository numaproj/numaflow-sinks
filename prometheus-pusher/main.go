package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/numaproj/numaflow-go/pkg/sink/server"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/zap"
)

const (
	PROMETHEUS_SERVER      = "PROMETHEUS_SERVER"
	SKIP_VALIDATION_FAILED = "SKIP_VALIDATION_FAILED"
)

type prometheusSink struct {
	logger     *zap.SugaredLogger
	skipFailed bool
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
	keys := make(map[string]bool)

	for _, payload := range msgPayloads {
		if _, ok := keys[payload.Name]; !ok {
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
				keys[payload.Name] = true
			}
			err = pusher.Push()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *prometheusSink) handle(ctx context.Context, datumList []sinksdk.Datum) sinksdk.Responses {
	ok := sinksdk.ResponsesBuilder()
	failed := sinksdk.ResponsesBuilder()
	var payloads []string
	for _, datum := range datumList {
		payloads = append(payloads, string(datum.Value()))
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
		failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to push the metrics"))
	}
	var pls []Payload
	for _, payloadMsg := range payloads {
		var pl Payload
		err := json.Unmarshal([]byte(payloadMsg), &pl)
		if !p.skipFailed && err != nil {
			return failed
		}
		pls = append(pls, pl)
	}
	err := p.push(pls)
	if err != nil {
		return failed
	}
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

func main() {
	logger := logging.NewLogger().Named("prometheus-sink")
	skipFailedStr := os.Getenv(SKIP_VALIDATION_FAILED)
	skipFailed := false
	var err error
	if skipFailedStr != "" {
		skipFailed, err = strconv.ParseBool(skipFailedStr)
		if err != nil {
			panic(err)
		}
	}
	ps := prometheusSink{logger: logger, skipFailed: skipFailed}
	server.New().RegisterSinker(sinksdk.SinkFunc(ps.handle)).Start(context.Background())
}
