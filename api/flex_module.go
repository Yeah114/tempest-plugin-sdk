package api

import (
	"context"
)

const NameFlexModule = "flex"

// FlexModule provides in-process cross-plugin communication primitives:
// - key/value store
// - topic pub/sub
// - request/response style API exposure and calls using JSON payloads ([]byte)
type FlexModule interface {
	Name() string

	Set(key string, val string)
	Get(key string) (string, bool)

	Publish(topic string, payloadJSON []byte)
	Subscribe(ctx context.Context, topic string) <-chan []byte

	Expose(apiName string, handler func(context.Context, []byte) ([]byte, string)) (func(), error)
	Call(ctx context.Context, apiName string, argsJSON []byte) ([]byte, string, error)
}
