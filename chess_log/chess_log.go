package main

import (
	"fmt"
	"os"
	"time"
	"github.com/jessevdk/go-flags"
	"github.com/kgigitdev/godgt"
	"encoding/json"
	"io/ioutil"
	"log"
)

var opts struct {
	Pngs bool   `long:"pngs" description:"Write PNG images of board updates"`
	Log bool	`long:"log" description:"Write Log output to console"`
	Port string `short:"p" long:"port" description:"Serial port" default:"/dev/ttyUSB0" env:"DGT_PORT"`
	Size int `short:"s" long:"size" description:"Image size" default:"128"`
	Filename string `short:"f" long:"filename" description:"File prefix for png image files" default:"boardupdate"`
}

type boardStatus struct {
	MoveCount   int       `json:"moveCount"`
	Timestamp   time.Time `json:"timestamp"`
	BoardUpdate string    `json:"boardUpdate"`
	ImageFile   string    `json:"imageFile"`
}

func main() {

	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		os.Exit(1)
	}

	dgtboard := godgt.NewDgtBoard(opts.Port)

	dgtboard.WriteCommand(godgt.DGT_SEND_RESET)
	dgtboard.WriteCommand(godgt.DGT_SEND_BRD)
	dgtboard.WriteCommand(godgt.DGT_SEND_UPDATE_BRD)

	go dgtboard.ReadLoop()
	var messageCount int
	var moveCount int

	startTime := time.Now().Format("2006-01-02T15.04.05Z")
	os.Mkdir(startTime, 0777)
	logPath := "./" + startTime + "/log.json"

	var statusArray []boardStatus

	for {
		select {
		case message := <-dgtboard.MessagesFromBoard:
			messageCount++
			filename := fmt.Sprintf("%s/%s-%04d.png",
				startTime, opts.Filename, moveCount)
			lastTurn, err := writeLog(message, filename, &moveCount)
			if err == 0 {
				statusArray = append(statusArray, lastTurn)
				sa, err := json.Marshal(statusArray)
				if err != nil {
					log.Fatal(err)
				} else {
					// print log to console if flag is set
					if opts.Log {
						lt, err := json.MarshalIndent(lastTurn, "", "\t")
						if err != nil {
							log.Fatal(err)
						} else {
							fmt.Println(string(lt))
						}
					}
					err := ioutil.WriteFile(logPath, sa, 0666)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
			if opts.Pngs && message.BoardUpdate != nil {
				fen := message.BoardUpdate.ToString()
				godgt.WritePng(fen, opts.Size, filename)
			}
			if message.FieldUpdate != nil {
				dgtboard.WriteCommand(godgt.DGT_SEND_BRD)
			}
		}
	}
}

func writeLog(m *godgt.Message, fileName string, mc *int) (boardStatus, int) {
	var currentStatus boardStatus
	if m.BoardUpdate != nil {
		*mc++
		currentStatus := boardStatus{
			*mc,
			time.Now(),
			m.ToString(),
			fileName,
		}
		return currentStatus, 0
	}
	return currentStatus, 1
}
