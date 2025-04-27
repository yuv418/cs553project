package main

import (
	abstraction "github.com/yuv418/cs553project/backend/common"
)

func main() {
	abstraction.AbsCtx.SetupMonolithDispatchTable()
	abstraction.AbsCtx.Run()
}
