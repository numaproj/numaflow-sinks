package main

import (
	"fmt"
	"strconv"
)

type PrometheusPayload struct {
	TimestampMs int64             `json:"timestampMs,omitempty"`
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Subsystem   string            `json:"subsystem,omitempty"`
	Type        string            `json:"type,omitempty"`
	Value       float64           `json:"value,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

func (p *PrometheusPayload) mergeLabels(labels map[string]string) {
	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	for key, val := range labels {
		// Should not override the payload label values
		if _, ok := p.Labels[key]; !ok {
			p.Labels[key] = val
		}

	}
}

func (p *PrometheusPayload) excludeLabels(labels []string) {
	if p.Labels == nil {
		return
	}
	for _, key := range labels {
		// Should not override the payload label values
		if _, ok := p.Labels[key]; ok {
			delete(p.Labels, key)
		}

	}
}

type OriginalPayload struct {
	UUID           string                 `json:"uuid"`
	ConfigID       string                 `json:"config_id"`
	CompositeKeys  []string               `json:"composite_keys"`
	Timestamp      int64                  `json:"timestamp"`
	UnifiedAnomaly any                    `json:"unified_anomaly"`
	Data           map[string]interface{} `json:"data"`
	Metadata       map[string]interface{} `json:"metadata"`
}

func (op *OriginalPayload) ConvertToPrometheusPayload(metricName string) *PrometheusPayload {

	value, err := strconv.ParseFloat(fmt.Sprintf("%.4f", op.UnifiedAnomaly), 64)
	if err != nil {
		value = 0
	}
	var labels map[string]string
	labels = make(map[string]string)
	for key, val := range op.Metadata {
		if key == "artifact_versions" {
			continue
		}
		if key == "model_version" {
			labels[key] = fmt.Sprintf("%.1f", val)

		} else {
			labels[key] = fmt.Sprintf("%s", val)
		}
	}
	namespace := op.Metadata["namespace"]
	if namespace == nil {
		namespace = ""
	}
	payload := &PrometheusPayload{
		Name:        metricName,
		TimestampMs: op.Timestamp,
		Namespace:   namespace.(string),
		Subsystem:   "none",
		Type:        "Gauge",
		Value:       value,
		Labels:      labels,
	}
	return payload

}
