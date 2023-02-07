package main

import (
	"crypto/tls"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttp_client(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

	}))
	hs := httpSink{}
	//creating http client
	client := &http.Client{Timeout: time.Duration(hs.timeout) * time.Second}
	if hs.skipInsecure {
		hs.logger.Info("Send insecure request")
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Timeout: 2 * time.Second, Transport: tr}
	}
	hs.httpClient = client
	hs.url = server.URL
	hs.method = http.MethodPost
	hs.logger = logging.NewLogger().Named("http-sink")
	err := hs.sendHTTPRequest(nil)
	assert.NoError(t, err)
}
