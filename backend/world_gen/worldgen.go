package worldgen

import (
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	"math/rand"
)

const PipesToGenerate int = 10000

func GenerateWorld(req *worldgenpb.WorldGenReq) worldgenpb.WorldGenerated {
	var pipeArray []*worldgenpb.PipeSpec

	// The game ID doesn't even matter. Maybe we can use it as a seed?
	gap := rand.Int31n(req.ViewportWidth / 3)

	for range PipesToGenerate {
		start := rand.Int31n(2 * (req.ViewportHeight / 3))
		height := rand.Int31n(((req.ViewportHeight - start) * 3) / 4)
		// Leave 1/4 of height for the pipe.
		pipeArray = append(pipeArray, &worldgenpb.PipeSpec{
			GapStart:  start,
			GapHeight: height,
		})
	}

	return worldgenpb.WorldGenerated{
		PipeSpacing: gap,
		PipeSpecs:   pipeArray,
	}

}
