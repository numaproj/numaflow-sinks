package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsPublisher struct {
	metricsTotalPushed      prometheus.Counter
	metricsTotalSuccess     prometheus.Counter
	metricsTotalFailed      prometheus.Counter
	metricsTotalSkipped     prometheus.Counter
	metricsAnomalyGenerated *prometheus.CounterVec
	labels                  map[string]string
}

func (mp *MetricsPublisher) registerMertics() {
	mp.metricsTotalPushed = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_metrics_pushed",
		Help:        "The total number of metrics pushed",
		ConstLabels: mp.labels,
	})
	mp.metricsTotalSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_metrics_success",
		Help:        "The total number of metrics successfully pushed",
		ConstLabels: mp.labels,
	})
	mp.metricsTotalFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_metrics_failed",
		Help:        "The total number of metrics failed push",
		ConstLabels: mp.labels,
	})
	mp.metricsTotalSkipped = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_metrics_skipped",
		Help:        "The total number of metrics skipped",
		ConstLabels: mp.labels,
	})

	mp.metricsAnomalyGenerated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "total_anomaly_generated",
		Help:        "The total count of anomaly score generator",
		ConstLabels: mp.labels,
	}, []string{"namespace", "app", "metrics"})
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

func (mp *MetricsPublisher) IncreaseAnomalyGenerated(namespace, app, metricName string) {
	mp.metricsAnomalyGenerated.WithLabelValues(namespace, app, metricName).Inc()
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
	return http.ListenAndServe(address, nil)
}
