//go:build initiator
// +build initiator

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupHandlers(ctx *abstraction.AbstractionServer) {
	SetupInitiatorHandler(ctx)
}
