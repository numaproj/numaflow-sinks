package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"net/http"
	"time"

	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/numaproj/numaflow-go/pkg/sink/server"
	flag2 "github.com/numaproj/numaflow-sinks/shared/flag"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

type httpSink struct {
	logger       *zap.SugaredLogger
	httpClient   *http.Client
	url          string
	method       string
	retries      int
	timeout      int
	windowing    int
	skipInsecure bool
	dropIfError  bool
	headers      arrayFlags
	metrics      *MetricsPublisher
}
type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (hs *httpSink) createHTTPClient() {
	//creating http client
	client := &http.Client{Timeout: time.Duration(hs.timeout) * time.Second}
	if hs.skipInsecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	hs.httpClient = client
}

func (hs *httpSink) sendHTTPRequest(data io.Reader) error {
	req, err := http.NewRequest(hs.method, hs.url, data)
	if err != nil {
		return err
	}
	if hs.httpClient == nil {
		return errors.New("HTTP Client is not initialized")
	}
	res, err := hs.httpClient.Do(req)
	if err != nil {
		return err
	}
	if res != nil {
		if res.Body != nil {
			res.Body.Close()
		}
		hs.logger.Infof("Response code: %d,", res.StatusCode)
	}
	return nil
}

func (hs *httpSink) handle(ctx context.Context, datumStreamCh <-chan sinksdk.Datum) sinksdk.Responses {
	ok := sinksdk.ResponsesBuilder()
	failed := sinksdk.ResponsesBuilder()
	for datum := range datumStreamCh {
		hs.metrics.IncreaseTotalCounter()
		hs.metrics.UpdateSize(float64(len(datum.Value())))
		//TODO Need to implemente parallel sending request
		data := bytes.NewReader(datum.Value())
		backoff := wait.Backoff{
			Steps:    hs.retries,
			Duration: 10 * time.Second,
			Factor:   2,
		}
		retryError := wait.ExponentialBackoffWithContext(ctx, backoff, func() (done bool, err error) {
			start := time.Now()
			err = hs.sendHTTPRequest(data)
			hs.metrics.UpdateLatency(float64(time.Since(start).Milliseconds()))
			if err != nil {
				hs.logger.Errorf("HTTP Request failed. %v", err)
				return false, nil
			}
			return true, nil
		})
		if retryError != nil {
			hs.logger.Errorf("HTTP Request failed. Error : %v", retryError)
			if hs.dropIfError {
				hs.metrics.IncreaseTotalDropped()
				hs.logger.Warn("Dropping messages due to failure")
				return ok
			}
			hs.metrics.IncreaseTotalFailed()
			failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to forward message"))
		}
		hs.metrics.IncreaseTotalSuccess()
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
	}
	if len(failed) > 0 {
		return failed
	}

	hs.logger.Info("HTTP Request send successfully")
	return ok
}

func main() {
	var metricPort int
	labels := flag2.MapFlag{}
	logger := logging.NewLogger().Named("http-sink")
	hs := httpSink{logger: logger}
	flag.StringVar(&hs.url, "url", "", "URL")
	flag.StringVar(&hs.method, "method", "GET", "HTTP Method")
	flag.IntVar(&hs.retries, "retries", 3, "Request Retries")
	flag.IntVar(&hs.timeout, "timeout", 30, "Request Timeout in seconds")
	flag.BoolVar(&hs.skipInsecure, "insecure", false, "Skip TLS verify")
	flag.BoolVar(&hs.dropIfError, "dropIfError", false, "Messages will drop after retry")
	flag.Var(&hs.headers, "headers", "HTTP Headers")
	flag.IntVar(&metricPort, "udsinkMetricsPort", 9090, "UDSink Metrics Port")
	flag.Var(&labels, "udsinkMetricsLabels", "UDSink Metrics Labels E.g: label=val1,label1=val2")
	// Parse the flag
	flag.Parse()

	hs.metrics = NewMetricsServer(labels)
	go hs.metrics.startMetricServer(metricPort)
	hs.logger.Infof("Metrics publisher initialized with port=%d", metricPort)
	//creating http client
	hs.createHTTPClient()
	hs.logger.Info("HTTP Sink starting successfully with args %v", hs)
	server.New().RegisterSinker(sinksdk.SinkFunc(hs.handle)).Start(context.Background())
}
