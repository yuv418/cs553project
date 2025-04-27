//go:build auth
// +build auth

package main

import abstraction "github.com/yuv418/cs553project/backend/common"

func SetupTables(ctx *abstraction.AbstractionServer) {
	SetupAuthTables(ctx)
}
