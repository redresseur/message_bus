package open_interface

type AuthorityHandler interface {
	CheckEndpoint(channelId string, point *EndPoint) error
}

type emptyAuthorityHandlerImpl struct {
}

func (*emptyAuthorityHandlerImpl) CheckEndpoint(channelId string, point *EndPoint) error {
	return nil
}

func NewEmptyAuthorityHandler() AuthorityHandler {
	return &emptyAuthorityHandlerImpl{}
}
