package worldgen

import (
	"log"
	"math/rand"

	"github.com/yuv418/cs553project/backend/commondata"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
)

const PipesToGenerate int = 10000

func GenerateWorld(ctx *commondata.ReqCtx, req *worldgenpb.WorldGenReq) (*worldgenpb.WorldGenerated, error) {
	var pipeArray []*worldgenpb.PipeSpec

	log.Printf("Got request params %v\n", req)

	// The game ID doesn't even matter. Maybe we can use it as a seed?
	log.Printf("req.ViewportWidth %d\n", req.ViewportWidth)
	gap := (req.ViewportWidth / 4) + rand.Int31n(req.ViewportWidth/4)

	for range PipesToGenerate {
		start := (req.ViewportHeight / 9) + rand.Int31n(2*(req.ViewportHeight/3))
		height := ((req.ViewportHeight - start) / 4) + rand.Int31n(((req.ViewportHeight - start) / 2))
		// Leave 1/4 of height for the pipe.
		pipeArray = append(pipeArray, &worldgenpb.PipeSpec{
			GapStart:  float64(start),
			GapHeight: float64(height),
		})
	}

	return &worldgenpb.WorldGenerated{
		PipeSpacing: float64(gap),
		PipeSpecs:   pipeArray,
	}, nil

}
