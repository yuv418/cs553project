package main

import (
	"log"

	abstraction "github.com/yuv418/cs553project/backend/common"
)

func main() {
	log.Printf("Microservices is set to %v\n", abstraction.AbsCtx.Microservice)

	SetupServiceData(abstraction.AbsCtx)
	SetupDispatchTable(abstraction.AbsCtx)
	SetupHandlers(abstraction.AbsCtx)

	abstraction.AbsCtx.Run()
}
