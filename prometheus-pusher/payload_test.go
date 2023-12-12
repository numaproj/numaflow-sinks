package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertToPrometheusPayload(t *testing.T) {
	JsonStr := `{"uuid":"35e0dc4603c845c9b999f5f669c64606","config_id":"test","composite_keys":["test_namespace","test_app","597b5bd8cc"],"timestamp":1701201827,"unified_anomaly":1.2,"data":{"namespace_app_rollouts_http_request_error_rate":null},"metadata":{"model_version":0,"artifact_versions":{"MinMaxScaler":"0","LSTMAE":"0","StdDevThreshold":"0"},"app":"test-app","intuit_alert":"true","namespace":"test-namespace","numalogic":"true","prometheus":"k8s-prometheus","rollouts_pod_template_hash":"597b5bd8cc"}}`
	var origiObj OriginalPayload
	err := json.Unmarshal([]byte(JsonStr), &origiObj)
	assert.NoError(t, err)
	prometheusPayload := origiObj.ConvertToPrometheusPayload("test")
	assert.Equal(t, prometheusPayload.TimestampMs, int64(1701201827))
	assert.Equal(t, prometheusPayload.Value, 1.2)
	assert.Equal(t, prometheusPayload.Labels["app"], "test-app")
	assert.Equal(t, prometheusPayload.Labels["namespace"], "test-namespace")
	assert.Equal(t, prometheusPayload.Labels["rollouts_pod_template_hash"], "597b5bd8cc")
}

func TestMergePrometheusLabelPayload(t *testing.T) {
	JsonStr := `{"uuid":"35e0dc4603c845c9b999f5f669c64606","config_id":"test","composite_keys":["test_namespace","test_app","597b5bd8cc"],"timestamp":1701201827,"unified_anomaly":1.2,"data":{"namespace_app_rollouts_http_request_error_rate":null},"metadata":{"model_version":0,"artifact_versions":{"MinMaxScaler":"0","LSTMAE":"0","StdDevThreshold":"0"},"app":"test-app","intuit_alert":"true","namespace":"test-namespace","numalogic":"true","prometheus":"k8s-prometheus","rollouts_pod_template_hash":"597b5bd8cc"}}`
	var origiObj OriginalPayload
	err := json.Unmarshal([]byte(JsonStr), &origiObj)
	assert.NoError(t, err)
	prometheusPayload := origiObj.ConvertToPrometheusPayload("test")
	labels := map[string]string{"label1": "value1", "label2": "value1"}
	prometheusPayload.mergeLabels(labels)
	assert.Equal(t, prometheusPayload.TimestampMs, int64(1701201827))
	assert.Equal(t, prometheusPayload.Value, 1.2)
	assert.Equal(t, prometheusPayload.Labels["app"], "test-app")
	assert.Equal(t, prometheusPayload.Labels["namespace"], "test-namespace")
	assert.Equal(t, prometheusPayload.Labels["rollouts_pod_template_hash"], "597b5bd8cc")
	assert.Equal(t, prometheusPayload.Labels["label1"], "value1")
}

func TestExcludePrometheusLabelPayload(t *testing.T) {
	JsonStr := `{"uuid":"35e0dc4603c845c9b999f5f669c64606","config_id":"test","composite_keys":["test_namespace","test_app","597b5bd8cc"],"timestamp":1701201827,"unified_anomaly":1.2,"data":{"namespace_app_rollouts_http_request_error_rate":null},"metadata":{"model_version":0,"artifact_versions":{"MinMaxScaler":"0","LSTMAE":"0","StdDevThreshold":"0"},"numalogic":"true", "app":"test-app","intuit_alert":"true","namespace":"test-namespace","numalogic":"true","prometheus":"k8s-prometheus","rollouts_pod_template_hash":"597b5bd8cc"}}`
	var origiObj OriginalPayload
	err := json.Unmarshal([]byte(JsonStr), &origiObj)
	assert.NoError(t, err)
	prometheusPayload := origiObj.ConvertToPrometheusPayload("test")
	labels := map[string]string{"label1": "value1", "label2": "value1"}
	exlabels := []string{"label1", "numalogic"}
	prometheusPayload.mergeLabels(labels)
	prometheusPayload.excludeLabels(exlabels)
	assert.Equal(t, prometheusPayload.TimestampMs, int64(1701201827))
	assert.Equal(t, prometheusPayload.Value, 1.2)
	assert.Equal(t, prometheusPayload.Labels["app"], "test-app")
	assert.Equal(t, prometheusPayload.Labels["namespace"], "test-namespace")
	assert.Equal(t, prometheusPayload.Labels["rollouts_pod_template_hash"], "597b5bd8cc")
	assert.NotContains(t, prometheusPayload.Labels, "label1")
	assert.NotContains(t, prometheusPayload.Labels, "numalogic")
}
