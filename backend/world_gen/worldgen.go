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
	var start int32

	log.Printf("Got request params %v\n", req)

	// The game ID doesn't even matter. Maybe we can use it as a seed?
	log.Printf("req.ViewportWidth %d\n", req.ViewportWidth)
	gap := (req.ViewportWidth / 4) + rand.Int31n(req.ViewportWidth/4)
	thresh := (2 * req.ViewportHeight) / 3
	maxHeight := (1 * req.ViewportHeight) / 3

	// Laziness
	prevClear := "up"

	for range PipesToGenerate {

		// If previously generated pipe is up, then can either generate center or up
		// If previously generated pipe is down, then can either generate center or down
		// If previously generated pipe is center, then can either generate anywhere
		if prevClear == "center" {
			// btw 1/9 and 7/9
			start = (req.ViewportHeight / 9) + rand.Int31n(thresh)
		} else if prevClear == "bottom" {
			// btw 4/9 and 7/9
			start = (4 * req.ViewportHeight / 9) + rand.Int31n(3*req.ViewportHeight/9)
		} else if prevClear == "up" {
			// btw 1/9 and 1/2
			start = (req.ViewportHeight / 9) + rand.Int31n(7*req.ViewportHeight/18)
		}

		if start < req.ViewportHeight/3 {
			prevClear = "up"
		} else if start > 2*req.ViewportHeight/3 {
			prevClear = "bottom"
		} else {
			prevClear = "center"
		}

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
