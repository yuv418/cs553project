//go:build worldgen
// +build worldgen

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupHandlers(ctx *abstraction.AbstractionServer) {
	SetupWorldgenHandler(ctx)
}
