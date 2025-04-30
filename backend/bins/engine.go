//go:build engine
// +build engine

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupTables(ctx *abstraction.AbstractionServer) {
	SetupGameEngineTables(ctx)
}
