package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type MetricsPublisher struct {
	metricsTotalPushed  prometheus.Counter
	metricsTotalSuccess prometheus.Counter
	metricsTotalFailed  prometheus.Counter
	metricsTotalSkipped prometheus.Counter
	labels              map[string]string
}

func (mp *MetricsPublisher) registerMertics() {
	mp.registerMetricsTotalPushed()
	mp.registerMetricsTotalSuccess()
	mp.registerMetricsTotalFailed()
	mp.registerMetricsTotalSkipped()

}

func (mp *MetricsPublisher) IncreaseTotalPushed() {
	mp.metricsTotalPushed.Inc()
}
func (mp *MetricsPublisher) IncreaseTotalSuccess() {
	mp.metricsTotalSuccess.Inc()
}
func (mp *MetricsPublisher) IncreaseTotalFailed() {
	mp.metricsTotalFailed.Inc()
}
func (mp *MetricsPublisher) IncreaseTotalSkipped() {
	mp.metricsTotalSkipped.Inc()
}

func (mp *MetricsPublisher) registerMetricsTotalPushed() {
	mp.metricsTotalPushed = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "metrics_total_pushed",
		Help:        "The total number of metrics pushed",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerMetricsTotalSuccess() {
	mp.metricsTotalSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "metrics_total_success",
		Help:        "The total number of metrics successfully pushed",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerMetricsTotalFailed() {
	mp.metricsTotalFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "metrics_total_failed",
		Help:        "The total number of metrics failed push",
		ConstLabels: mp.labels,
	})
}

func (mp *MetricsPublisher) registerMetricsTotalSkipped() {
	mp.metricsTotalSkipped = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "metrics_total_skipped",
		Help:        "The total number of metrics skipped",
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
