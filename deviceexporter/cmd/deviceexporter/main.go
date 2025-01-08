package main

import (
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/info"
	"os"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/deviceexporter/internal/app"

	"github.com/sirupsen/logrus"
)

func main() {
	a := app.NewApp(info.GetVersionString())
	if err := a.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
