package main

import (
	abstraction "github.com/yuv418/cs553project/backend/common"
)

func main() {
	srv := abstraction.NewAbstractionServer()
	srv.SetupMonolithDispatchTable()
	srv.Run()
}
