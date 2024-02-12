package main

import (
	"context"
	"encoding/json"
	"fmt"
	cache "github.com/go-pkgz/expirable-cache/v2"
	"github.com/numaproj/numaflow-go/pkg/mapper"
	log "github.com/numaproj/numaflow/pkg/shared/logging"
	"go.uber.org/zap"
	"time"
)

// Forward is a mapper that directly forward the input to the output
type BackFill struct {
	cache        cache.Cache[string, []map[string]interface{}]
	windowLength int
	log          *zap.SugaredLogger
}

func (f *BackFill) processPayload(payload *Payload) {
	if payload.Metadata["role"] == "stable" {
		key := getCacheKey(payload.Metadata)
		f.log.Infof("caching Stable data, key=%s", key)
		f.cache.Set(key, payload.Data, 0)
	} else if payload.Metadata["role"] == "canary" {
		if len(payload.Data) < f.windowLength {
			key := getCacheKey(payload.Metadata)
			stableData, found := f.cache.Get(key)
			if !found {
				return
			}
			f.log.Infof("original data=%v", payload.Data)
			payload.Data = f.backFill(stableData, payload.Data)

			f.log.Infof("canary data update from stable data, key=%s", key)
			f.log.Infof("updated data=%v", payload.Data)
		}
	}
}

func (f *BackFill) backFill(stable []map[string]interface{}, canary []map[string]interface{}) []map[string]interface{} {
	backFillLength := f.windowLength - len(canary)
	backFillElements := stable[f.windowLength-backFillLength:]
	return append(backFillElements, canary...)
}

func getCacheKey(metadata map[string]interface{}) string {
	namespace := metadata["namespace"]
	app := metadata["app"]
	return fmt.Sprintf("%s-%s", namespace, app)
}

func (f *BackFill) Map(ctx context.Context, keys []string, d mapper.Datum) mapper.Messages {
	// directly forward the input to the output
	messages := mapper.MessagesBuilder()
	var payload Payload

	err := json.Unmarshal(d.Value(), &payload)
	if err != nil {
		f.log.Warnf("unmarshalling error %v", err)
		messages.Append(mapper.MessageToDrop())
	}

	//for _, paylond := range payloads {
	f.processPayload(&payload)
	payload.Metadata["numalogic_opex_tags"] = map[string]string{
		"source": "numalogic-rollouts",
	}
	if len(payload.Data) < f.windowLength {
		f.log.Warnf("uuid= %s dropping message because not enough data. metadata:%v, data length: %d", payload.UUID, payload.Metadata, len(payload.Data))
		return mapper.MessagesBuilder().Append(mapper.MessageToDrop())
	}
	//}
	resultVal, err := json.Marshal(payload)
	if err != nil {
		f.log.Warnf("uuid= %s payload is dropping  because not enough data. metadata=%v", payload.UUID, payload.Metadata)
		return mapper.MessagesBuilder().Append(mapper.MessageToDrop())
	}
	var resultKeys = keys
	return mapper.MessagesBuilder().Append(mapper.NewMessage(resultVal).WithKeys(resultKeys))
}

func main() {
	cache1 := cache.NewCache[string, []map[string]interface{}]().WithTTL(10 * time.Minute)
	err := mapper.NewServer(&BackFill{windowLength: 12, log: log.NewLogger(), cache: cache1}).Start(context.Background())
	if err != nil {
		panic("Failed to start map function server: " + err.Error())
	}
}
