package main

import (
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/info"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/kpsmi/internal/app"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	a := app.NewApp(info.GetVersionString())
	if err := a.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
