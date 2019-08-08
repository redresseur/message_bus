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

// 绑定死亡动作
func (k *KeepAlive)BindDeathFunc(deathFunc DeathFunc)  {
	k.deathFunc = deathFunc
}

func (k *KeepAlive)Update()  {
	k.deadLine = time.Now().Add(k.heartBeatDuration)
}

func (k *KeepAlive)Run(){
	ticker := time.NewTicker(k.heartBeatDuration)
	defer ticker.Stop()

	for  {
		// 如果是超时了
		if k.deadLine.Before(time.Now()){
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

