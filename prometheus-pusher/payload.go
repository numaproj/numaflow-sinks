package main

type Payload struct {
	TimestampMs int64             `json:"timestampMs,omitempty"`
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Subsystem   string            `json:"subsystem,omitempty"`
	Type        string            `json:"type,omitempty"`
	Value       float64           `json:"value,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}
