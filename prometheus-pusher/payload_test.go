package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertToPrometheusPayload(t *testing.T) {
	JsonStr := `{"uuid":"35e0dc4603c845c9b999f5f669c64606","config_id":"test","composite_keys":["test_namespace","test_app","597b5bd8cc"],"timestamp":1701201827,"unified_anomaly":1.2,"data":{"namespace_app_rollouts_cpu_utilization":0.517299409015888,"namespace_app_rollouts_http_request_error_rate":0.517299409015888,"namespace_app_rollouts_memory_utilization":0.517299409015888,"namespace_app_rollouts_http_requests_latency":0.517299409015888},"metadata":{"model_version":0,"artifact_versions":{"MinMaxScaler":"0","LSTMAE":"0","StdDevThreshold":"0"},"app":"test-app","intuit_alert":"true","namespace":"test-namespace","numalogic":"true","prometheus":"k8s-prometheus","rollouts_pod_template_hash":"597b5bd8cc"}}`
	var origiObj OriginalPayload
	err := json.Unmarshal([]byte(JsonStr), &origiObj)
	assert.NoError(t, err)
	prometheusPayload := origiObj.ConvertToPrometheusPayload("test")
	assert.Equal(t, prometheusPayload[0].TimestampMs, int64(1701201827))
	assert.Equal(t, prometheusPayload[0].Value, 1.2)
	assert.Equal(t, prometheusPayload[0].Labels["app"], "test-app")
	assert.Equal(t, prometheusPayload[0].Labels["namespace"], "test-namespace")
	assert.Equal(t, prometheusPayload[0].Labels["rollouts_pod_template_hash"], "597b5bd8cc")

	payloadJson, err := json.Marshal(prometheusPayload)
	assert.NoError(t, err)
	assert.Equal(t, prometheusPayload[1].TimestampMs, int64(1701201827))
	assert.Equal(t, 0.5173, prometheusPayload[1].Value)
	assert.Contains(t, string(payloadJson), "namespace_app_rollouts_cpu_utilization_anomaly")
	assert.Contains(t, string(payloadJson), "namespace_app_rollouts_memory_utilization_anomaly")
	assert.Contains(t, string(payloadJson), "namespace_app_rollouts_http_request_error_rate_anomaly")
	assert.Contains(t, string(payloadJson), "namespace_app_rollouts_http_requests_latency_anomaly")
	assert.Equal(t, prometheusPayload[1].Labels["app"], "test-app")
	assert.Equal(t, prometheusPayload[1].Labels["namespace"], "test-namespace")
	assert.Equal(t, prometheusPayload[1].Labels["rollouts_pod_template_hash"], "597b5bd8cc")
}
