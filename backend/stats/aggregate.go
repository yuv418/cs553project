package stats

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/yuv418/cs553project/backend/commondata"
)

type Stat struct {
	SrcSvcName  string
	SrcSvcVerb  string
	DestSvcName string
	DestSvcVerb string
	GameId      string
	// This could be latency,
	// or something else, who knows
	ReqTime time.Duration
}

// https://gobyexample.com/channels
func StartStatThread() chan *Stat {
	messageChan := make(chan *Stat)

	statDir := commondata.GetEnv("STAT_DIR", "statout")
	os.MkdirAll(statDir, 0755)

	statFile := filepath.Join(statDir, "stats.csv")
	file, err := os.Create(statFile)
	if err != nil {
		log.Fatalf("Failed at creating %s\n", statFile)
	}
	writer := csv.NewWriter(file)
	dataArray := []string{"SrcSvcName", "SrcSvcVerb", "DestSvcName", "DestSvcVerb", "GameId", "ReqTime"}

	writer.Write(dataArray)
	writer.Flush()

	go (func() {
		defer file.Close()

		for statEnt := range messageChan {
			dataArray[0] = statEnt.SrcSvcName
			dataArray[1] = statEnt.SrcSvcVerb
			dataArray[2] = statEnt.DestSvcName
			dataArray[3] = statEnt.DestSvcVerb
			dataArray[4] = statEnt.GameId
			dataArray[5] = statEnt.ReqTime.String()

			writer.Write(dataArray)
			writer.Flush()
		}
	})()

	return messageChan
}
