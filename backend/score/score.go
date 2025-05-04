package score

// https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson
// https://pkg.go.dev/os#File.Write
// https://pkg.go.dev/encoding/json

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/yuv418/cs553project/backend/common"
	"github.com/yuv418/cs553project/backend/commondata"
	scorepb "github.com/yuv418/cs553project/backend/protos/score"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ScoreCtx struct {
	jsonFile *os.File
	// game id -> entry
	data map[string][]*scorepb.ScoreEntry
}

func LoadScoreCtx() (*ScoreCtx, error) {
	scoreFileName := common.GetEnv("SCORE_FILE", "score.json")
	scoreFile, err := os.OpenFile(scoreFileName, os.O_RDWR, 0)
	if err != nil && os.IsNotExist(err) {
		// Doesn't exist, try to make it
		scoreFile, err = os.Create(scoreFileName)
		if err != nil {
			return nil, err
		}

		return &ScoreCtx{
			jsonFile: scoreFile,
			data:     make(map[string][]*scorepb.ScoreEntry),
		}, nil
	} else if err == nil {
		// Exists it, read
		ctx := &ScoreCtx{
			jsonFile: scoreFile,
		}

		data, err := io.ReadAll(ctx.jsonFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &ctx.data)
		if err != nil {
			return nil, err
		}

		return ctx, nil
	} else {
		// Some other error
		return nil, err
	}

}

func (ctx *ScoreCtx) UpdateScore(reqCtx *commondata.ReqCtx, req *scorepb.ScoreEntry) (*empty.Empty, error) {
	log.Printf("(UpdateScore) Received request for %v\n", req)

	// Set
	ctx.data[reqCtx.Username] = append(ctx.data[reqCtx.Username], req)

	// Write to file
	out, err := json.Marshal(ctx.data)
	if err != nil {
		return nil, err
	}

	ctx.jsonFile.Truncate(0)
	ctx.jsonFile.Seek(0, 0)

	// TODO check bytes write
	_, err = ctx.jsonFile.Write(out)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (ctx *ScoreCtx) GetScores(reqCtx *commondata.ReqCtx, _ *emptypb.Empty) (*scorepb.GetScoresResp, error) {
	log.Printf("(GetScores) Received request for %s\n", reqCtx.Username)

	return &scorepb.GetScoresResp{
		Entries: ctx.data[reqCtx.Username],
	}, nil
}
