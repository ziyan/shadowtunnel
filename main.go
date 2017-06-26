package main

import (
	"os"
	"runtime"

	"github.com/ziyan/shadowtunnel/cli"
)

func main() {
	// use all CPU cores for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	cli.Run(os.Args)
}
