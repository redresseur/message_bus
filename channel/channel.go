package channel

import (
	"context"
	"errors"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/storage"
	"sync"
)

type channelHandlerImpl struct {
	ctx context.Context
	// 在内存中緩存所有channel信息
	// TODO: 換爲redis
	contexts sync.Map

	storageHandler open_interface.StorageHandler
}

func NewChannelHandler(ctx context.Context) open_interface.ChannelHandler {
	return &channelHandlerImpl{
		ctx:            ctx,
		storageHandler: &storage.StorageHandlerImpl{},
	}
}

func (ch *channelHandlerImpl) ExitChannel(channelId string, point *open_interface.EndPoint) error {
	cc, ok := ch.contexts.Load(channelId)
	if !ok {
		return open_interface.ErrChannelNotExisted
	}

	chCtx := cc.(*open_interface.ChannelContext)
	chCtx.RemoveEndPoint(point.Id)

	return nil
}

func (ch *channelHandlerImpl) ChildrenChannel(parentContext *open_interface.ChannelContext,
	childInfo *open_interface.ChannelInfo, point *open_interface.EndPoint) (*open_interface.ChannelContext, error) {
	// TODO: 待實現
	return nil, errors.New("implement me")
}

func (ch *channelHandlerImpl) Channel(channelId string) *open_interface.ChannelContext {
	cc, ok := ch.contexts.Load(channelId)
	if !ok {
		return nil
	}

	return cc.(*open_interface.ChannelContext)
}

func (ch *channelHandlerImpl) CreateChannel(info *open_interface.ChannelInfo, point *open_interface.EndPoint) (*open_interface.ChannelContext, error) {
	if _, ok := ch.contexts.Load(info.ChannelId); ok {
		return nil, open_interface.ErrChannelHasBeenExisted
	}

	st, _ := ch.storageHandler.New(info.ChannelId)
	cc := open_interface.WithChannelContext(ch.ctx, info, st)

	point.IsCreator = true
	cc.AddEndPoint(point)

	ch.contexts.Store(info.ChannelId, cc)
	return cc, nil
}

func (ch *channelHandlerImpl) JoinChannel(channelId string, point *open_interface.EndPoint) error {
	cc, ok := ch.contexts.Load(channelId)
	if !ok {
		return open_interface.ErrChannelNotExisted
	}

	return cc.(*open_interface.ChannelContext).AddEndPoint(point)
}

func (ch *channelHandlerImpl) CloseChannel(channelId string, point *open_interface.EndPoint) error {
	cc, ok := ch.contexts.Load(channelId)
	if !ok {
		return nil
	}

	// 检查身份
	chCtx := cc.(*open_interface.ChannelContext)
	if ep := chCtx.EndPoint(point.Id); ep != nil {
		return open_interface.ErrEndPointNotExisted
	} else if !ep.IsCreator {
		return open_interface.ErrIsNotCreator
	}

	chCtx.Cancel()
	return nil
}

func (ch *channelHandlerImpl) ListenChannel(channelId string, point *open_interface.EndPoint) (context.Context, error) {
	cc, ok := ch.contexts.Load(channelId)
	if !ok {
		return nil, open_interface.ErrEndPointNotExisted
	}

	ctx, err := cc.(*open_interface.ChannelContext).BindRW(point.Id, point.RW)
	if err != nil {
		return nil, err
	}

	// <-ctx.Done()
	return ctx, nil
}
