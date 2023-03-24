package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/argoproj/argo-workflows/v3/util/template"
	"github.com/numaproj/numaflow/pkg/shared/logging"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/client-go/rest"

	numaexpr "github.com/numaproj/numaflow-sinks/argoworkflow/shared/expr"

	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/numaproj/numaflow-go/pkg/sink/server"
)

const (
	ARGO_WORKFLOW_TEMPLATE   = "ARGO_WORKFLOW_TEMPLATE"
	WORKFLOW_NAMESPACE       = "WORKFLOW_NAMESPACE"
	PARAMETER_NAME           = "PARAMETER_NAME"
	WORKFLOW_SERVICE_ACCOUNT = "WORKFLOW_SERVICE_ACCOUNT"
	MSG_DEDUP_KEYS           = "MSG_DEDUP_KEYS"
	DEDUP_CACHE_LIMIT        = "DEDUP_CACHE_LIMIT"
	DEDUP_CACHE_TTL_DURATION = "DEDUP_CACHE_TTL_DURATION"
	READ_INTERVAL_DURATION   = "READ_INTERVAL_DURATION"
	WORKFLOW_NAME_PREFIX     = "WORKFLOW_NAME_PREFIX"
)

const wf = `{
  "apiVersion": "argoproj.io/v1alpha1",
  "kind": "Workflow",
  "metadata": {
    "generateName": "{{WORKFLOW_NAME_PREFIX}}-wf-",
    "namespace": "{{WORKFLOW_NAMESPACE}}"
  },
  "spec": {
    "serviceAccountName": "{{WORKFLOW_SERVICE_ACCOUNT}}",
    "workflowTemplateRef": {
      "name": "{{ARGO_WORKFLOW_TEMPLATE}}"
    }
  }
}`

type argoWfSink struct {
	dedupKeys    []string
	logger       *zap.SugaredLogger
	envMap       map[string]string
	wfObj        *v1alpha1.Workflow
	wfClientSet  wfclientset.Interface
	cacheKeys    *cache.LRUExpireCache
	keyTTL       time.Duration
	readInterval time.Duration
}

func (as *argoWfSink) submitWorkflow(param string) error {
	ctx := context.Background()
	defer ctx.Done()
	wf := as.wfObj.DeepCopy()

	args := v1alpha1.Arguments{
		Parameters: []v1alpha1.Parameter{
			{
				Name:  as.envMap[PARAMETER_NAME],
				Value: v1alpha1.AnyStringPtr(param),
			},
		},
	}
	wf.Spec.Arguments = args
	created, err := as.wfClientSet.ArgoprojV1alpha1().Workflows(as.envMap[WORKFLOW_NAMESPACE]).Create(ctx, wf, v1.CreateOptions{})

	if err != nil {
		return err
	}
	as.logger.Infof("%s Workflow created successful ", created.Name)
	return nil
}

func (as *argoWfSink) handle(ctx context.Context, datumStreamCh <-chan sinksdk.Datum) sinksdk.Responses {
	ok := sinksdk.ResponsesBuilder()
	failed := sinksdk.ResponsesBuilder()
	var payloads []string

	for datum := range datumStreamCh {
		if len(as.dedupKeys) > 0 {
			key, err := as.getKey(datum.Value())
			if err != nil {
				failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), err.Error()))
			}
			if _, ok := as.cacheKeys.Get(key); !ok {
				payloads = append(payloads, string(datum.Value()))
				as.cacheKeys.Add(key, true, as.keyTTL)
				as.logger.Infof("Added unique key %s", key)
			} else {
				as.logger.Infof("Duplicate %s message skipped", key)
			}
		} else {
			payloads = append(payloads, string(datum.Value()))
		}
		ok = ok.Append(sinksdk.ResponseOK(datum.ID()))
		failed = failed.Append(sinksdk.ResponseFailure(datum.ID(), "failed to trigger workflow"))
	}
	if len(payloads) == 0 {
		return ok
	}
	params, err := json.Marshal(payloads)
	if err != nil {
		as.logger.Errorf("Payload marshal failed. %v", err)
		return failed
	}
	err = as.submitWorkflow(string(params))
	if err != nil {
		as.logger.Errorf("Workflow submission failed. %v", err)
		return failed
	}
	// read Interval
	time.Sleep(as.readInterval)
	return ok
}

func (as *argoWfSink) constructKeyExpression(keys []string) string {
	var expr bytes.Buffer
	for _, key := range keys {
		expr.WriteString(fmt.Sprintf("%s%s,'-',", numaexpr.JsonRoot, strings.TrimSpace(key)))
	}
	result := expr.String()
	if result == "" {
		return ""
	}
	return fmt.Sprintf("sprig.nospace(sprig.cat(%s))", result[:len(result)-5])
}

func (as *argoWfSink) getKey(msg []byte) (string, error) {
	key, err := numaexpr.EvalExpression(as.constructKeyExpression(as.dedupKeys), msg)
	if err != nil {
		return "", err
	}
	return key.(string), err
}

func getNamespace() string {
	if ns, ok := os.LookupEnv("WORKFLOW_NAMESPACE"); ok {
		return ns
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return "default"
}

func initialize(kubeConfig *rest.Config) (*argoWfSink, error) {
	logger := logging.NewLogger().Named("argo-workflow-sink")
	env := make(map[string]string)
	wftmp, found := os.LookupEnv(ARGO_WORKFLOW_TEMPLATE)
	if !found {
		return nil, fmt.Errorf("%s env not found", ARGO_WORKFLOW_TEMPLATE)
	}
	env[ARGO_WORKFLOW_TEMPLATE] = wftmp
	env[WORKFLOW_NAMESPACE] = getNamespace()
	paramName, found := os.LookupEnv(PARAMETER_NAME)
	if !found {
		return nil, fmt.Errorf("%s env not found", PARAMETER_NAME)
	}
	env[PARAMETER_NAME] = paramName
	env[WORKFLOW_SERVICE_ACCOUNT] = os.Getenv(WORKFLOW_SERVICE_ACCOUNT)
	prefix, found := os.LookupEnv(WORKFLOW_NAME_PREFIX)
	if !found {
		env[WORKFLOW_NAME_PREFIX] = "argo-"
	}
	env[WORKFLOW_NAME_PREFIX] = prefix
	dedupKeyStr := os.Getenv(MSG_DEDUP_KEYS)
	var dedupKeys []string
	if dedupKeyStr != "" {
		dedupKeys = strings.Split(dedupKeyStr, ",")
	}
	wfClientSet := wfclientset.NewForConfigOrDie(kubeConfig)
	resultWfStr, err := template.Replace(wf, env, true)
	if err != nil {
		return nil, err
	}
	wfObj := v1alpha1.Workflow{}
	err = json.Unmarshal([]byte(resultWfStr), &wfObj)
	if err != nil {
		return nil, err
	}

	dedupLimitStr, found := os.LookupEnv(DEDUP_CACHE_LIMIT)
	if !found {
		dedupLimitStr = "100"
	}
	dedupLimit, err := strconv.ParseInt(dedupLimitStr, 10, 0)
	if err != nil {
		return nil, err
	}

	ttlDurationStr, found := os.LookupEnv(DEDUP_CACHE_TTL_DURATION)
	if !found {
		ttlDurationStr = "1m"
	}
	ttlDuration, err := time.ParseDuration(ttlDurationStr)

	if err != nil {
		return nil, err
	}

	intervalDurationStr, found := os.LookupEnv(READ_INTERVAL_DURATION)
	if !found {
		intervalDurationStr = "1m"
	}
	intervalDuration, err := time.ParseDuration(intervalDurationStr)

	if err != nil {
		return nil, err
	}

	cache := cache.NewLRUExpireCache(int(dedupLimit))

	logger.Info("Argo workflow sink initialized successful")
	return &argoWfSink{envMap: env, wfClientSet: wfClientSet, wfObj: &wfObj, logger: logger, dedupKeys: dedupKeys, cacheKeys: cache, keyTTL: ttlDuration, readInterval: intervalDuration}, nil
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	wfSink, err := initialize(config)
	if err != nil {
		panic(err)
	}
	//sinksdk.Start(context.Background(), wfSink.handle)
	server.New().RegisterSinker(sinksdk.SinkFunc(wfSink.handle)).Start(context.Background())

}
