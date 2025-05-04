//go:build monolith
// +build monolith

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupHandlers(ctx *abstraction.AbstractionServer) {
	SetupAuthHandler(ctx)
	SetupWorldgenHandler(ctx)
	SetupGameEngineHandler(ctx)
	SetupInitiatorHandler(ctx)
	SetupMusicHandler(ctx)
	SetupScoreHandler(ctx)
}
