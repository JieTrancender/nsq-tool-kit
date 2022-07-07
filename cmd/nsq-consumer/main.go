package main

import (
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	nsqconsumer.NewApp("nsq-tool-kit").Run()
}
