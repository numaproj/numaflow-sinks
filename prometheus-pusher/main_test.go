package main

import (
	"encoding/json"
	"fmt"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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
		param := "key1=,key2=value2"
		data := parseStringToMap(param)
		assert.Len(t, data, 2)
		assert.Equal(t, data["key1"], "")
		assert.Equal(t, data["key2"], "value2")
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

func TestPusher(t *testing.T) {
	// Fake a Pushgateway that responds with 202 to DELETE and with 200 in
	// all other cases.

	var metrics []io_prometheus_client.MetricFamily

	pgwOK := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dec := expfmt.NewDecoder(r.Body, expfmt.FmtProtoDelim)

			var mf io_prometheus_client.MetricFamily
			dec.Decode(&mf)
			metrics = append(metrics, mf)
			fmt.Println(mf)
			w.Header().Set("Content-Type", `text/plain; charset=utf-8`)
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer pgwOK.Close()
	logger := logging.NewLogger().Named("prometheus-sink")
	ps := prometheusSink{logger: logger, skipFailed: false, labels: nil}
	payloadMsg := `[{"TimestampMs":1680124991883,"name":"namespace_app_rollouts_unified_anomaly","namespace":"dev_devx_o11yfuzzygqlfederation_usw2_stg","type":"Gauge","value":0.4944,"labels":{"app":"o11y_fuzzy_gql_federation","intuit_alert":"true","model_version":"1","namespace":"dev_devx_o11yfuzzygqlfederation_usw2_stg","rollouts_pod_template_hash":"794dcbf4b7"}},
					{"TimestampMs":1680124991883,"name":"namespace_app_rollouts_unified_anomaly","namespace":"dev_devx_o11yfuzzygqlfederation_usw2_stg","type":"Gauge","value":0.4955,"labels":{"app":"o11y_fuzzy_gql_federation","intuit_alert":"true","model_version":"1","namespace":"dev_devx_o11yfuzzygqlfederation_usw2_stg","rollouts_pod_template_hash":"794dcbf4b33"}}]`
	var pl []Payload
	json.Unmarshal([]byte(payloadMsg), &pl)
	t.Setenv("PROMETHEUS_SERVER", pgwOK.URL)
	ps.metrics = &MetricsPublisher{}
	ps.metrics.metricsAnomalyGenerated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "total_anomaly_generated_1",
		Help:        "The total count of anomaly score generator",
		ConstLabels: nil,
	}, []string{"namespace", "app", "metrics"})
	ps.metrics.metricsTotalSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "total_success_pused_1",
		Help:        "The total count of anomaly score generator",
		ConstLabels: nil,
	})
	defer func() { ps.metrics.metricsAnomalyGenerated = nil }()
	ps.push(pl)

	assert.Equal(t, pl[0].Name, metrics[0].GetName())
	assert.Equal(t, strings.ToUpper(pl[0].Type), metrics[0].Type.String())
	assert.Equal(t, pl[0].Value, *metrics[0].GetMetric()[0].Gauge.Value)

	assert.Equal(t, pl[1].Name, metrics[1].GetName())
	assert.Equal(t, strings.ToUpper(pl[1].Type), metrics[1].Type.String())
	assert.Equal(t, pl[1].Value, *metrics[1].GetMetric()[0].Gauge.Value)
}
