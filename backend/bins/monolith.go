//go:build monolith
// +build monolith

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupTables(ctx *abstraction.AbstractionServer) {
	SetupAuthTables(ctx)
	SetupWorldgenTables(ctx)
	SetupGameEngineTables(ctx)
	SetupInitiatorTables(ctx)
	SetupMusicTables(ctx)
}
