package commondata

import "context"

type ReqCtx struct {
	HttpCtx  *context.Context
	Username string
	Jwt      string
}
