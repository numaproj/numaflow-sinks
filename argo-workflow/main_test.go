package main

import (
	"context"
	"encoding/json"
	"fmt"
	fakewfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/fake"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sink"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"testing"
	"time"
)

type message struct {
	value string
}

func (m message) Value() []byte {
	return []byte(m.value)
}
func (m message) EventTime() time.Time {
	return time.Now()
}
func (m message) Watermark() time.Time {
	return time.Now()
}
func (m message) ID() string {
	return ""
}

func TestWorkflow_Submission(t *testing.T) {

	t.Setenv("KUBERNETES_SERVICE_HOST", "localhost")
	t.Setenv("KUBERNETES_SERVICE_PORT", "8080")
	t.Setenv(ARGO_WORKFLOW_TEMPLATE, "test")
	t.Setenv(PARAMETER_NAME, "param1")
	t.Setenv(WORKFLOW_NAMESPACE, "default")
	t.Setenv(WORKFLOW_SERVICE_ACCOUNT, "default")
	t.Setenv(READ_INTERVAL_DURATION, "0")
	//argoSink := argoWfSink{wfClientSet: wfclientset, envMap: map[string]string{ARGO_WORKFLOW_TEMPLATE: "test", WORKFLOW_NAMESPACE: "default", PARAMETER_NAME: "param"}}
	t.Run("Without Dedu key", func(t *testing.T) {
		wfclientset := fakewfclientset.NewSimpleClientset()
		config := rest.Config{}
		argoSink, err := initialize(&config)
		assert.NoError(t, err)
		argoSink.wfClientSet = wfclientset

		msgs := []sinksdk.Datum{
			message{value: "test"},
		}
		_ = argoSink.handle(context.Background(), msgs)
		wflist, err := wfclientset.ArgoprojV1alpha1().Workflows("default").List(context.Background(), v1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, wflist.Items, 1)
		fmt.Println(wflist.Items[0].Name)
	})
	t.Run("With dedup key", func(t *testing.T) {
		wfclientset := fakewfclientset.NewSimpleClientset()
		t.Setenv(MSG_DEDUP_KEYS, ".keys,.name,.bu")
		config := rest.Config{}
		argoSink, err := initialize(&config)
		assert.NoError(t, err)
		argoSink.wfClientSet = wfclientset
		payload := `{ "keys" :["1","welcome","3","4","5"], "name":"test", "namespace":"default", "bu":"001", "subApp":{"name":"test", "bu":"devx"}}`
		msgs := []sinksdk.Datum{
			message{value: payload},
			message{value: payload},
		}
		_ = argoSink.handle(context.Background(), msgs)
		wflist, err := wfclientset.ArgoprojV1alpha1().Workflows("default").List(context.Background(), v1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, wflist.Items, 1)
		assert.Len(t, wflist.Items[0].Spec.Arguments.Parameters, 1)
		params, err := json.Marshal([]string{payload})
		assert.NoError(t, err)
		assert.Equal(t, string(params), wflist.Items[0].Spec.Arguments.Parameters[0].Value.String())

	})
	t.Run("With dedup key", func(t *testing.T) {
		wfclientset := fakewfclientset.NewSimpleClientset()
		t.Setenv(MSG_DEDUP_KEYS, ".keys,.name,.bu")
		config := rest.Config{}
		argoSink, err := initialize(&config)
		assert.NoError(t, err)
		argoSink.wfClientSet = wfclientset
		payload := `{ "keys" :["1","welcome","3","4","5"], "name":"test", "namespace":"default", "bu":"001", "subApp":{"name":"test", "bu":"devx"}}`
		msgs := []sinksdk.Datum{
			message{payload},
			message{payload},
		}
		_ = argoSink.handle(context.Background(), msgs)
		wflist, err := wfclientset.ArgoprojV1alpha1().Workflows("default").List(context.Background(), v1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, wflist.Items, 1)
		assert.Len(t, wflist.Items[0].Spec.Arguments.Parameters, 1)
		params, err := json.Marshal([]string{payload})
		assert.NoError(t, err)
		assert.Equal(t, string(params), wflist.Items[0].Spec.Arguments.Parameters[0].Value.String())

	})
	t.Run("With dedup key with catch", func(t *testing.T) {
		wfclientset := fakewfclientset.NewSimpleClientset()
		t.Setenv(MSG_DEDUP_KEYS, ".keys,.name,.bu")
		t.Setenv(DEDUP_CACHE_LIMIT, "2")
		t.Setenv(DEDUP_CACHE_TTL_DURATION, "2s")
		t.Setenv(READ_INTERVAL_DURATION, "1s")
		config := rest.Config{}
		argoSink, err := initialize(&config)
		assert.NoError(t, err)
		argoSink.wfClientSet = wfclientset
		payload := `{ "keys" :["1","welcome","3","4","5"], "name":"test", "namespace":"default", "bu":"001", "subApp":{"name":"test", "bu":"devx"}}`
		msgs := []sinksdk.Datum{
			message{payload},
			message{payload},
		}
		_ = argoSink.handle(context.Background(), msgs)
		assert.NoError(t, err)
		wflist, err := wfclientset.ArgoprojV1alpha1().Workflows("default").List(context.Background(), v1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, wflist.Items, 1)
		assert.Len(t, wflist.Items[0].Spec.Arguments.Parameters, 1)
		params, err := json.Marshal([]string{payload})
		assert.NoError(t, err)
		assert.Equal(t, string(params), wflist.Items[0].Spec.Arguments.Parameters[0].Value.String())
		_ = argoSink.handle(context.Background(), msgs)
		assert.NoError(t, err)
		wflist, err = wfclientset.ArgoprojV1alpha1().Workflows("default").List(context.Background(), v1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, wflist.Items, 1)
		assert.Len(t, wflist.Items[0].Spec.Arguments.Parameters, 1)
		params, err = json.Marshal([]string{payload})
		assert.NoError(t, err)
	})

}

func TestConstructKeyExpression(t *testing.T) {
	config := rest.Config{}
	t.Setenv(ARGO_WORKFLOW_TEMPLATE, "test")
	t.Setenv(PARAMETER_NAME, "param1")
	t.Setenv(WORKFLOW_NAMESPACE, "default")
	t.Setenv(WORKFLOW_SERVICE_ACCOUNT, "default")
	argoSink, err := initialize(&config)
	assert.NoError(t, err)
	expr := argoSink.constructKeyExpression(nil)
	assert.Equal(t, "", expr)
}
