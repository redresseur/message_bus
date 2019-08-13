package open_interface

import (
	"context"
	"time"
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
}

func NewKeeper(ctx context.Context, heartBeatDuration time.Duration, deathFunc DeathFunc) *KeepAlive {
	res := &KeepAlive{
		deadLine:  time.Now().Add(heartBeatDuration),
		ctx:       ctx,
		deathFunc: deathFunc,
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
	k.deadLine = time.Now().Add(k.heartBeatDuration)
}

func (k *KeepAlive) Finish() {
	k.ctx.Value(k).(context.CancelFunc)()
}

func (k *KeepAlive) Run() {
	ticker := time.NewTicker(k.heartBeatDuration)
	defer ticker.Stop()

	for {
		// 如果是超时了
		if k.deadLine.Before(time.Now()) {
			k.deathFunc()
			break
		}

		select {
		case <-ticker.C:
			continue
		case <-k.ctx.Done():
			return
		}
	}

	return
}
