package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type MetricsPublisher struct {
	payloadTotalCounter prometheus.Counter
	payloadTotalSuccess prometheus.Counter
	payloadTotalFailed  prometheus.Counter
	payloadTotalDropped prometheus.Counter
	payloadLatency      prometheus.Summary
	payloadSize         prometheus.Summary
	labels              map[string]string
}

func (mp *MetricsPublisher) registerMertics() {
	mp.registerPayloadTotalCounter()
	mp.registerPayloadTotalSuccess()
	mp.registerPayloadTotalFailed()
	mp.registerPayloadTotalDropped()
	mp.registerPayloadSize()
	mp.registerPayloadLatency()

}

func (mp *MetricsPublisher) IncreaseTotalCounter() {
	mp.payloadTotalCounter.Inc()
}

func (mp *MetricsPublisher) IncreaseTotalSuccess() {
	mp.payloadTotalSuccess.Inc()
}
func (mp *MetricsPublisher) IncreaseTotalFailed() {
	mp.payloadTotalFailed.Inc()
}
func (mp *MetricsPublisher) IncreaseTotalDropped() {
	mp.payloadTotalDropped.Inc()
}

func (mp *MetricsPublisher) UpdateSize(size float64) {
	mp.payloadSize.Observe(size)
}

func (mp *MetricsPublisher) UpdateLatency(latency float64) {
	mp.payloadLatency.Observe(latency)
}

func (mp *MetricsPublisher) registerPayloadTotalCounter() {
	mp.payloadTotalCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "payload_total_count",
		Help:        "The total number of payload events",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerPayloadTotalSuccess() {
	mp.payloadTotalSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "payload_total_success",
		Help:        "The total number of success payload events",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerPayloadTotalFailed() {
	mp.payloadTotalFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "payload_total_failed",
		Help:        "The total number of failed payload events",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerPayloadTotalDropped() {
	mp.payloadTotalDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "payload_total_dropped",
		Help:        "The total number of dropped payload events",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerPayloadLatency() {
	mp.payloadLatency = promauto.NewSummary(prometheus.SummaryOpts{
		Name:        "payload_send_latency",
		Help:        "The payload round trip duration",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerPayloadSize() {
	mp.payloadSize = promauto.NewSummary(prometheus.SummaryOpts{
		Name:        "payload_size",
		Help:        "payload size",
		ConstLabels: mp.labels,
	})
}
func NewMetricsServer(labels map[string]string) *MetricsPublisher {
	metricsPublisher := &MetricsPublisher{}
	metricsPublisher.labels = labels
	metricsPublisher.registerMertics()
	return metricsPublisher

}
func (mp *MetricsPublisher) startMetricServer(port int) error {
	address := fmt.Sprintf(":%d", port)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(address, nil)
	return nil
}
