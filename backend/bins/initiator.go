//go:build initiator
// +build initiator

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupTables(ctx *abstraction.AbstractionServer) {
	SetupInitiatorTables(ctx)
}
