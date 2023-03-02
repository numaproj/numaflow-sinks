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
	mp.payloadTotalCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_request_count",
		Help:        "The total number of payload events",
		ConstLabels: mp.labels,
	})
	mp.payloadTotalSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_request_success",
		Help:        "The total number of success payload events",
		ConstLabels: mp.labels,
	})
	mp.payloadTotalFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_request_failed",
		Help:        "The total number of failed payload events",
		ConstLabels: mp.labels,
	})
	mp.payloadTotalDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_request_dropped",
		Help:        "The total number of dropped payload events",
		ConstLabels: mp.labels,
	})
	mp.payloadLatency = promauto.NewSummary(prometheus.SummaryOpts{
		Name:        "total_request_latency",
		Help:        "The payload round trip duration",
		ConstLabels: mp.labels,
	})
	mp.payloadSize = promauto.NewSummary(prometheus.SummaryOpts{
		Name:        "total_request_size",
		Help:        "total request size",
		ConstLabels: mp.labels,
	})
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
