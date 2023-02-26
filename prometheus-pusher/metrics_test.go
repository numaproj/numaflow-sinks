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
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalPushed))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalSuccess))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalFailed))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.metricsTotalSkipped))
	mp.IncreaseTotalPushed()
	mp.IncreaseTotalSuccess()
	mp.IncreaseTotalSkipped()
	mp.IncreaseTotalFailed()
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalPushed))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalSuccess))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalFailed))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.metricsTotalSkipped))
}
