package channel

import (
	"context"
	"errors"
	"github.com/redresseur/message_bus/definitions"
	"sync"
)


type channelHandlerImpl struct {
	ctx context.Context
	// 在内存中緩存所有channel信息
	// TODO: 換爲redis
	contexts sync.Map
}

func NewChannelHandler(ctx context.Context) definitions.ChannelHandler  {
	return &channelHandlerImpl{
		ctx:ctx,
	}
}

func (ch *channelHandlerImpl) ExitChannel(channelId string, point *definitions.EndPoint) error {
	cc, ok := ch.contexts.Load(channelId)
	if ! ok {
		return definitions.ErrChannelNotExisted
	}

	chCtx := cc.(*definitions.ChannelContext)
	chCtx.RemoveEndPoint(point.Id)

	return nil
}

func (ch *channelHandlerImpl) ChildrenChannel(parentContext *definitions.ChannelContext,
	childInfo *definitions.ChannelInfo, point *definitions.EndPoint) (*definitions.ChannelContext, error) {
	// TODO: 待實現
	return  nil, errors.New("implement me")
}

func (ch *channelHandlerImpl) Channel(channelId string) *definitions.ChannelContext {
	cc, _ := ch.contexts.Load(channelId)
	return cc.(*definitions.ChannelContext)
}

func (ch *channelHandlerImpl) CreateChannel(info *definitions.ChannelInfo, point *definitions.EndPoint) (*definitions.ChannelContext, error) {
	if _, ok := ch.contexts.Load(info.ChannelId); ok {
		return nil, definitions.ErrChannelHasBeenExisted
	}

	cc := definitions.WithChannelContext(ch.ctx, info)

	point.IsCreator = true
	return cc, cc.AddEndPoint(point)
}

func (ch *channelHandlerImpl) JoinChannel(channelId string, point *definitions.EndPoint) error {
	cc , ok := ch.contexts.Load(channelId)
	if ! ok{
		return definitions.ErrChannelNotExisted
	}

	return cc.(*definitions.ChannelContext).AddEndPoint(point)
}

func (ch *channelHandlerImpl) CloseChannel(channelId string, point *definitions.EndPoint) error {
	cc , ok := ch.contexts.Load(channelId)
	if ! ok{
		return nil
	}

	// 检查身份
	chCtx := cc.(*definitions.ChannelContext)
	if ep := chCtx.EndPoint(point.Id); ep != nil{
		return definitions.ErrEndPointNotExisted
	}else if !ep.IsCreator{
		return definitions.ErrIsNotCreator
	}

	chCtx.Cancel()
	return nil
}

func (ch *channelHandlerImpl) ListenChannel(channelId string, point *definitions.EndPoint)(context.Context, error) {
	cc , ok := ch.contexts.Load(channelId)
	if ! ok{
		return nil, definitions.ErrEndPointNotExisted
	}

	ctx, err := cc.(*definitions.ChannelContext).BindRW(point.Id, point.RW)
	if err != nil{
		return nil, err
	}

	// <-ctx.Done()
	return ctx, nil
}

