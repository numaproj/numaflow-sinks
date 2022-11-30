package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/numaproj/numaflow-go/pkg/sink/server"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type httpSink struct {
	logger       *zap.SugaredLogger
	url          string
	method       string
	retries      int
	timeout      int
	windowing    int
	skipInsecure bool
	headers      arrayFlags
}
type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (hs *httpSink) sendHTTPRequest(data io.Reader) error {
	client := &http.Client{Timeout: time.Duration(hs.timeout) * time.Second}
	if hs.skipInsecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Timeout: 2 * time.Second, Transport: tr}
	}
	req, err := http.NewRequest(hs.method, hs.url, data)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res != nil {
		hs.logger.Infof("Response code: %d, Response Body: %s", res.StatusCode, res.Body)
	}
	return nil
}

func (hs *httpSink) handle(ctx context.Context, datumList []sinksdk.Datum) sinksdk.Responses {
	ok := sinksdk.ResponsesBuilder()
	failed := sinksdk.ResponsesBuilder()
	var payloads []string
	for _, datum := range datumList {
		payloads = append(payloads, string(datum.Value()))
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
		failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to trigger workflow"))
	}
	payloadBytes, err := json.Marshal(payloads)
	if err != nil {
		hs.logger.Errorf("Payload json marshall failed. %v", err)
		return failed
	}
	data := bytes.NewReader(payloadBytes)
	err = hs.sendHTTPRequest(data)
	if err != nil {
		hs.logger.Errorf("HTTP Request failed. %v", err)
		return failed
	}
	hs.logger.Info("HTTP Request send successfully")
	return ok
}

func main() {
	logger := logging.NewLogger().Named("http-sink")
	hs := httpSink{logger: logger}
	flag.StringVar(&hs.url, "url", "", "URL")
	flag.StringVar(&hs.method, "method", "GET", "HTTP Method")
	flag.IntVar(&hs.retries, "retries", 3, "Request Retries")
	flag.IntVar(&hs.timeout, "timeout", 30, "Request Timeout in seconds")
	flag.BoolVar(&hs.skipInsecure, "insecure-skip-tls-verify", false, "Skip TLS verify")
	flag.Var(&hs.headers, "headers", "HTTP Headers")

	// Parse the flag
	flag.Parse()

	server.New().RegisterSinker(sinksdk.SinkFunc(hs.handle)).Start(context.Background())
}
