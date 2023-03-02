package main

import (
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsPublisher(t *testing.T) {
	mp := NewMetricsServer(map[string]string{"label": "val1", "label2": "val2"})
	mp.IncreaseTotalCounter()
	mp.IncreaseTotalSuccess()
	mp.IncreaseTotalDropped()
	mp.IncreaseTotalFailed()
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.payloadTotalCounter))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.payloadTotalDropped))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.payloadTotalSuccess))
	assert.Equal(t, float64(1), testutil.ToFloat64(mp.payloadTotalFailed))
	mp.IncreaseTotalCounter()
	mp.IncreaseTotalSuccess()
	mp.IncreaseTotalDropped()
	mp.IncreaseTotalFailed()
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.payloadTotalCounter))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.payloadTotalDropped))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.payloadTotalSuccess))
	assert.Equal(t, float64(2), testutil.ToFloat64(mp.payloadTotalFailed))
}
