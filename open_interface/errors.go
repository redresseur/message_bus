package open_interface

import "errors"

var (
	ErrEndPointHasBeenExisted = errors.New("the end point has been existed")
	ErrEndPointNotExisted = errors.New("the end point is not existed in channel")
	ErrChannelHasBeenExisted = errors.New("the channel has been existed")
	ErrChannelNotExisted = errors.New("the channel is not existed")
	ErrIsNotCreator = errors.New("the end point is not creator")
	ErrKeyIsInvalid = errors.New("the key is invalid in storage")
)