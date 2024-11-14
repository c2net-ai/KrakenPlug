package main

import (
	"openi.pcl.ac.cn/Kraken/KrakenPlug/kprunc/internal/rt"
	"os"
)

func main() {
	rt := rt.NewRuntime()
	err := rt.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
