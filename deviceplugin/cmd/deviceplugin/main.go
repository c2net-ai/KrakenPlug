package main

import (
	"os"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/deviceplugin/internal/app"

	"github.com/sirupsen/logrus"
)

var (
	BuildVersion = "Filled by the build system"
)

func main() {
	a := app.NewApp(BuildVersion)
	if err := a.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
