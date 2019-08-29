package context

import (
	"context"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestMultiValueContext_Value(t *testing.T) {
	multiCtx, err := WithMultiValueContext(context.Background(), "key" )
	if ! assert.EqualValues(t, ErrKVNotPair ,err, "should be ErrKVNotPair") {
		t.SkipNow()
	}

	multiCtx, err = WithMultiValueContext(context.Background(), "key", "value" )
	if ! assert.NoError(t, err) {
		t.SkipNow()
	}

	if ! assert.NoError(t, UpdateMultiValueContext(multiCtx, "key", "value")){
		t.SkipNow()
	}

	assert.EqualValues(t, "value" , multiCtx.Value("key"), "value not match")

}

func TestUpdateMultiValueContext(t *testing.T) {
	multiCtx, err := WithMultiValueContext(context.Background(), HeartBeat, nil)
	if ! assert.NoError(t, err) {
		t.SkipNow()
	}

	UpdateMultiValueContext(multiCtx, HeartBeat, open_interface.ErrHeartBeatTimeOut)
	assert.Equal(t, open_interface.ErrHeartBeatTimeOut, multiCtx.Value(HeartBeat))
}