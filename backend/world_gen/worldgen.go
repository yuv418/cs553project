package worldgen

import (
	"log"
	"math/rand"

	"github.com/yuv418/cs553project/backend/commondata"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
)

const PipesToGenerate int = 100

func GenerateWorld(ctx *commondata.ReqCtx, req *worldgenpb.WorldGenReq) (*worldgenpb.WorldGenerated, error) {
	var pipeArray []*worldgenpb.PipeSpec
	var height int32

	log.Printf("Got request params %v\n", req)

	// The game ID doesn't even matter. Maybe we can use it as a seed?
	log.Printf("req.ViewportWidth %d\n", req.ViewportWidth)
	gap := (req.ViewportWidth / 4) + rand.Int31n(req.ViewportWidth/4)
	thresh := (6 * req.ViewportHeight) / 9
	maxHeight := (1 * req.ViewportHeight) / 3

	for range PipesToGenerate {
		start := (req.ViewportHeight / 9) + rand.Int31n(thresh)
		remaining := req.ViewportHeight - start
		// If the remaining amount is less than the gap
		if remaining < thresh {
			height = ((2 * remaining) / 3) + rand.Int31n(remaining/6)
		} else if remaining > maxHeight {
			height = ((1 * remaining) / 4) + rand.Int31n(remaining/4)
		} else {
			height = (remaining / 2) + rand.Int31n(remaining/2)
		}

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
