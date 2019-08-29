package context

import (
	"context"
	"errors"
	"math"
)

var (
	ErrKVNotPair            = errors.New("the key-value pair is not invalid")
	ErrNotMultiValueContext = errors.New("the context type is not multi_value context")
)

type MultiValueContext struct {
	context.Context
	kvPairs map[interface{}]interface{}
}

func (mc *MultiValueContext) Value(key interface{}) interface{} {
	if v, ok := mc.kvPairs[key]; ok {
		return v
	}

	return nil
}

// 一个value ctx 中存储多个 k-v 键值对
func WithMultiValueContext(parentCtx context.Context, kv ...interface{}) (context.Context, error) {
	multiCtx := &MultiValueContext{
		Context: parentCtx,
		kvPairs: map[interface{}]interface{}{},
	}
	kvLen := len(kv)
	if kvLen == 0 || math.Mod(float64(kvLen), 2) != 0 {
		return nil, ErrKVNotPair
	}

	for i := 0; i < kvLen; {
		multiCtx.kvPairs[kv[i]] = kv[i+1]
		i += 2
	}

	return multiCtx, nil
}

func UpdateMultiValueContext(ctx context.Context, k, v interface{}) error {
	multiCtx, ok := ctx.(*MultiValueContext)
	if !ok {
		return ErrNotMultiValueContext
	}

	multiCtx.kvPairs[k] = v
	return nil
}
