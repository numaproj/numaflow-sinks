package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"io"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"time"

	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/numaproj/numaflow-go/pkg/sink/server"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"go.uber.org/zap"
)

type httpSink struct {
	logger       *zap.SugaredLogger
	url          string
	method       string
	retries      int
	timeout      int
	windowing    int
	skipInsecure bool
	dropIfError  bool
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
		hs.logger.Info("Send insecure request")
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
	for _, datum := range datumList {
		//TODO Need to implemente parallel sending request
		data := bytes.NewReader(datum.Value())
		backoff := wait.Backoff{
			Steps:    hs.retries,
			Duration: 10 * time.Second,
			Factor:   2,
		}
		retryError := wait.ExponentialBackoffWithContext(ctx, backoff, func() (done bool, err error) {
			err = hs.sendHTTPRequest(data)
			if err != nil {
				hs.logger.Errorf("HTTP Request failed. %v", err)
				return false, nil
			}
			return true, nil
		})
		if retryError != nil {
			hs.logger.Errorf("HTTP Request failed. Error : %v", retryError)
			if hs.dropIfError {
				hs.logger.Warn("Dropping messages due to failure")
				return ok
			}
			failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to forward message"))
		}
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
	}
	if len(failed) > 0 {
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
	flag.BoolVar(&hs.skipInsecure, "insecure", false, "Skip TLS verify")
	flag.BoolVar(&hs.dropIfError, "dropIfError", false, "Messages will drop after retry")
	flag.Var(&hs.headers, "headers", "HTTP Headers")

	// Parse the flag
	flag.Parse()

	hs.logger.Info("HTTP Sink starting successfully with args %v", hs)
	server.New().RegisterSinker(sinksdk.SinkFunc(hs.handle)).Start(context.Background())
}
