package main

import (
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttp_client(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

	}))
	hs := httpSink{}
	hs.url = server.URL
	hs.method = http.MethodPost
	hs.logger = logging.NewLogger().Named("http-sink")
	err := hs.sendHTTPRequest(nil)
	assert.Error(t, err)

	hs.createHTTPClient()
	err = hs.sendHTTPRequest(nil)
	assert.NoError(t, err)
}
