package open_interface

import (
	"context"
	"errors"
	"time"
)

const MinHeartBeatDuration = 3 * time.Second

var (
	ErrHeartBeatTimeOut = errors.New("heart beating was time out")
)

type DeathFunc func()

type KeepAlive struct {
	// 死亡时间
	deadLine time.Time

	// 上下文
	ctx context.Context

	// 间隔时间
	heartBeatDuration time.Duration

	// 心跳结束时执行的动作
	deathFunc DeathFunc

	updateSignal chan struct{}
}

func NewKeeper(ctx context.Context, heartBeatDuration time.Duration, deathFunc DeathFunc) *KeepAlive {
	if heartBeatDuration < MinHeartBeatDuration {
		heartBeatDuration = MinHeartBeatDuration
	}

	res := &KeepAlive{
		deadLine:          time.Now().Add(heartBeatDuration),
		ctx:               ctx,
		deathFunc:         deathFunc,
		heartBeatDuration: heartBeatDuration,
		updateSignal:      make(chan struct{}, 1024),
	}

	newCtx, cancel := context.WithCancel(ctx)
	res.ctx = context.WithValue(newCtx, res, cancel)

	return res
}

// 绑定死亡动作
func (k *KeepAlive) BindDeathFunc(deathFunc DeathFunc) {
	k.deathFunc = deathFunc
}

func (k *KeepAlive) Update() {
	defer func() {
		recover()
	}()
	k.updateSignal <- struct{}{}
}

func (k *KeepAlive) Finish() {
	k.ctx.Value(k).(context.CancelFunc)()
	close(k.updateSignal)
}

func (k *KeepAlive) Run() {
	// 每隔3 s 轮询一次
	ticker := time.NewTicker(MinHeartBeatDuration)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-k.updateSignal:
			{
				if !ok {
					return
				} else {
					k.deadLine = time.Now().Add(k.heartBeatDuration)
				}
			}
		case <-ticker.C:
			{
				// 如果是超时了
				if k.deadLine.Before(time.Now()) {
					k.deathFunc()
					return
				}
			}
		case <-k.ctx.Done():
			return
		}
	}

	return
}
