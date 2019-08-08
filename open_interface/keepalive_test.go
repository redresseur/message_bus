package open_interface

import (
	"context"
	"testing"
	"time"
)

func TestKeepAlive_Run(t *testing.T) {
	heartBeatDuration := 20*time.Second
	keeper := &KeepAlive{
		deadLine: time.Now().Add(heartBeatDuration),
		ctx: context.Background(),
		heartBeatDuration: heartBeatDuration,
		deathFunc: func() {
			t.Logf("hearbeat over")
		},
	}

	go func() {
		time.Sleep(heartBeatDuration/2)
		keeper.Update()
	}()

	keeper.Run()
}
