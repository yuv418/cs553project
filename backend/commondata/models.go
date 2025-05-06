package commondata

import (
	"bufio"
	"context"
)

type ReqCtx struct {
	HttpCtx  *context.Context
	Username string
	Jwt      string
	GameId   string

	TargetSvcName string
	TargetSvcVerb string
}

type WebTransportHandle struct {
	WtStream any
	Writer   *bufio.Writer
}
