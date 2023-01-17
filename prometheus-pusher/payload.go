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

func (p *Payload) mergeLabels(labels map[string]string) {
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

func (p *Payload) excludeLabels(labels []string) {
	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	for _, key := range labels {
		if _, ok := p.Labels[key]; ok {
			delete(p.Labels, key)
		}

	}
}
