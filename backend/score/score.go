package score

// https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson
// https://pkg.go.dev/os#File.Write
// https://pkg.go.dev/encoding/json

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/emirpasic/gods/trees/binaryheap"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/yuv418/cs553project/backend/common"
	"github.com/yuv418/cs553project/backend/commondata"
	scorepb "github.com/yuv418/cs553project/backend/protos/score"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ScoreCtx struct {
	jsonFile *os.File
	// game id -> entry
	data       map[string][]*scorepb.ScoreEntry
	globalHeap *binaryheap.Heap
}

func LoadScoreCtx() (*ScoreCtx, error) {
	scoreFileName := common.GetEnv("SCORE_FILE", "score.json")
	scoreFile, err := os.OpenFile(scoreFileName, os.O_RDWR, 0)

	ctx := &ScoreCtx{}
	ctx.globalHeap = binaryheap.NewWith(func(a, b interface{}) int {
		// Max heap
		return int(b.(*scorepb.ScoreEntry).Score - a.(*scorepb.ScoreEntry).Score)
	})

	if err != nil && os.IsNotExist(err) {
		// Doesn't exist, try to make it
		scoreFile, err = os.Create(scoreFileName)
		if err != nil {
			return nil, err
		}

		ctx.jsonFile = scoreFile
		ctx.data = make(map[string][]*scorepb.ScoreEntry)

		// Initial write
		err := ctx.WriteScores()
		if err != nil {
			return nil, err
		}

		return ctx, nil

	} else if err == nil {
		// Exists it, read
		ctx.jsonFile = scoreFile

		data, err := io.ReadAll(ctx.jsonFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &ctx.data)
		if err != nil {
			return nil, err
		}

		// Load ordered data

		// https://bitfieldconsulting.com/posts/map-iteration
		for k, entries := range ctx.data {
			for _, entry := range entries {
				entry.Username = &k
				ctx.globalHeap.Push(entry)
			}
		}

		return ctx, nil
	} else {
		// Some other error
		return nil, err
	}

}

func (ctx *ScoreCtx) WriteScores() error {
	// Write to file
	out, err := json.Marshal(ctx.data)
	if err != nil {
		return err
	}

	ctx.jsonFile.Truncate(0)
	ctx.jsonFile.Seek(0, 0)

	// TODO check bytes write
	_, err = ctx.jsonFile.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *ScoreCtx) UpdateScore(reqCtx *commondata.ReqCtx, req *scorepb.ScoreEntry) (*empty.Empty, error) {
	log.Printf("(UpdateScore) Received request for %v\n", req)

	// Set
	ctx.data[reqCtx.Username] = append(ctx.data[reqCtx.Username], req)

	// Write
	err := ctx.WriteScores()
	if err != nil {
		return nil, err
	}

	req.Username = &reqCtx.Username

	// Note there may be a bug here?
	ctx.globalHeap.Push(req)

	return &emptypb.Empty{}, nil
}

func (ctx *ScoreCtx) GetScores(reqCtx *commondata.ReqCtx, _ *emptypb.Empty) (*scorepb.GetScoresResp, error) {
	log.Printf("(GetScores) Received request for %s\n", reqCtx.Username)

	i := 0
	it := ctx.globalHeap.Iterator()
	// I don't like the efficiency of this
	globalEntries := make([]*scorepb.ScoreEntry, 0, 5)

	// https://stackoverflow.com/questions/21950244/is-there-a-way-to-iterate-over-a-range-of-integers
	for it.Next() && i < 5 {
		globalEntries = append(globalEntries, it.Value().(*scorepb.ScoreEntry))
		i++
	}

	return &scorepb.GetScoresResp{
		Entries:       ctx.data[reqCtx.Username],
		GlobalEntries: globalEntries,
	}, nil
}
