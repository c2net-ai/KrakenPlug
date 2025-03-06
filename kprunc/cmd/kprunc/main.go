package main

import (
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/kprunc/internal/rt"
	"os"
)

func main() {
	rt := rt.NewRuntime()
	err := rt.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
