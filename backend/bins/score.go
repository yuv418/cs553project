//go:build score
// +build score

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupTables(ctx *abstraction.AbstractionServer) {
	SetupScoreTables(ctx)
}
