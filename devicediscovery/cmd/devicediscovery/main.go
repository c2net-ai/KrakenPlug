package main

import (
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/info"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/devicediscovery/internal/app"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	a := app.NewApp(info.GetVersionString())
	if err := a.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
