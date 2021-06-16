package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"

	"simple-golang-crawler/engine"
	"simple-golang-crawler/parser"
	"simple-golang-crawler/persist"
	"simple-golang-crawler/scheduler"

	"github.com/alexflint/go-arg"
)

func init() {
	log.SetFormatter(&nested.Formatter{
		FieldsOrder: []string{
			log.FieldKeyTime, log.FieldKeyLevel, log.FieldKeyFile,
			log.FieldKeyFunc, log.FieldKeyMsg,
		},
		CustomCallerFormatter: func(f *runtime.Frame) string {
			return fmt.Sprintf("(%s %s:%d)", f.Function, path.Base(f.File), f.Line)
		},
		TimestampFormat:  time.RFC3339,
		HideKeys:         true,
		NoFieldsSpace:    true,
		NoUppercaseLevel: true,
		TrimMessages:     true,
		CallerFirst:      true,
	})
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
}

var cmdArgs struct {
	IdType    string `arg:"-t,--type" help:"id type,support:\n\taid:\t\t下载单个视频\n\tupid\t\t下载指定up主的视频"`
	Id        int64  `arg:"-i,--id" help:"视频或up主的id"`
	WorkCount int    `arg:"-w,--worker" help:"并行数"`
}

func main() {
	var err error
	var idType string
	var id int64
	arg.MustParse(&cmdArgs)
	if cmdArgs.IdType == "" {
		fmt.Println("Please enter your id type(`aid` or `upid`)")
		fmt.Scan(&idType)
		fmt.Println("Please enter your id")
		fmt.Scan(&id)
	} else {
		idType = cmdArgs.IdType
		id = cmdArgs.Id
	}

	if cmdArgs.WorkCount == 0 {
		cmdArgs.WorkCount = 30
	}

	var req *engine.Request
	if idType == "aid" {
		req = parser.GetRequestByAid(id)
	} else if idType == "upid" {
		req = parser.GetRequestByUpId(id)
	} else {
		log.Fatalln("Wrong type you enter")
		os.Exit(1)
	}

	itemProcessFun := persist.GetItemProcessFun()
	var wg sync.WaitGroup
	wg.Add(1)
	itemChan, err := itemProcessFun(&wg)
	if err != nil {
		panic(err)
	}

	queueScheduler := scheduler.NewConcurrentScheduler()
	conEngine := engine.NewConcurrentEngine(cmdArgs.WorkCount, queueScheduler, itemChan)
	log.Println("Start working.")
	conEngine.Run(req)
	wg.Wait()
	log.Println("All work has done")
}
