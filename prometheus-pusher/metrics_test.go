package main

import (
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsPublisher(t *testing.T) {
	mp := NewMetricsServer(map[string]string{"label": "val1", "label2": "val2"})
	mp.IncreaseTotalPushed()
	mp.IncreaseTotalSuccess()
	mp.IncreaseTotalSkipped()
	mp.IncreaseTotalFailed()
	mp.IncreaseAnomalyGenerated("test", "app1", "anomaly1")
	mp.IncreaseAnomalyGenerated("test1", "app2", "anomaly1")
	mp.IncreaseAnomalyGenerated("test", "app1", "anomaly1")
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalPushed))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalSuccess))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalFailed))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalSkipped))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsAnomalyGenerated.WithLabelValues("test", "app1", "anomaly1")))
	assert.Equal(t, float64(0), testutil.ToFloat64(mp.metricsAnomalyGenerated.WithLabelValues("test1", "app1", "anomaly1")))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsAnomalyGenerated.WithLabelValues("test1", "app2", "anomaly1")))
	mp.IncreaseTotalPushed()
	mp.IncreaseTotalSuccess()
	mp.IncreaseTotalSkipped()
	mp.IncreaseTotalFailed()
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalPushed))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalSuccess))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalFailed))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalSkipped))
}
