package main

type Payload struct {
	UUID     string                   `json:"uuid"`
	ConfigID string                   `json:"config_id"`
	Data     []map[string]interface{} `json:"data"`
	Startts  int64                    `json:"start_time"`
	Endts    int64                    `json:"end_time"`
	Metadata map[string]interface{}   `json:"metadata"`
}
