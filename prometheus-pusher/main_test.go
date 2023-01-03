package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseStringToMap(t *testing.T) {
	t.Run("Good_value", func(t *testing.T) {
		param := "key=value1,key1=value2,key2=value3"
		data := parseStringToMap(param)
		assert.Len(t, data, 3)
		assert.Equal(t, data["key"], "value1")
		assert.Equal(t, data["key1"], "value2")
		assert.Equal(t, data["key2"], "value3")
	})
	t.Run("No_value", func(t *testing.T) {
		param := ""
		data := parseStringToMap(param)
		assert.Len(t, data, 0)
	})
	t.Run("Invalid_value1", func(t *testing.T) {
		param := "key=value1,key1"
		data := parseStringToMap(param)
		assert.Len(t, data, 1)
	})
	t.Run("Invalid_value2", func(t *testing.T) {
		param := "key1=,key2=val"
		data := parseStringToMap(param)
		assert.Len(t, data, 2)
	})
}

func TestMergeLabel(t *testing.T) {
	payloadMsg := `{"namespace": "dev-devx-o11yfuzzygqlfederation-usw2-qal", "name": "namespace_http_numalogic_o11yfuzzygqlfederation_segment_api_error_count_anomaly_anomaly_anomaly_anomaly_anomaly_anomaly_anomaly_anomaly_anomaly", "model_config": "default", "resume_training": true}`
	var pl Payload
	param := "key=value1,key1=value2,key2=value3"
	data := parseStringToMap(param)
	err := json.Unmarshal([]byte(payloadMsg), &pl)
	assert.NoError(t, err)
	assert.Len(t, pl.Labels, 0)
	pl.mergeLabels(data)
	assert.Len(t, pl.Labels, 3)

}
