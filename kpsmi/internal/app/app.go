package app

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
	"strings"
)

func NewApp(buildVersion ...string) *cli.App {
	c := cli.NewApp()
	c.Name = "kpsmi"
	c.Usage = ""
	if len(buildVersion) == 0 {
		buildVersion = append(buildVersion, "")
	}
	c.Version = buildVersion[0]

	c.Flags = []cli.Flag{}

	c.Action = func(c *cli.Context) error {
		return action(c)
	}

	return c
}

func action(c *cli.Context) (err error) {
	printInfo()
	return nil
}

func printInfo() {
	err, stdout, stderr, close := utils.MaskPrint()
	if err != nil {
		return
	}
	defer close()

	t := table.NewWriter()
	style := table.StyleDefault
	style.Format.Header = text.FormatDefault
	t.SetStyle(style)

	t.SetTitle("KPSMI")

	t.AppendHeader(table.Row{"Card", "Vendor", "Model", "Memory-Usage(MB)", "Util(%)"})

	var count int
	d, err := util.NewDevice()
	if err != nil {
		goto exit
	}
	defer d.Shutdown()

	count, err = d.GetDeviceCount()
	for i := 0; i < count; i++ {
		memInfo, err := d.GetDeviceMemoryInfo(i)
		if err != nil {
			continue
		}
		model, err := d.GetDeviceModel(i)
		if err != nil {
			continue
		}
		util, err := d.GetDeviceUtil(i)
		if err != nil {
			continue
		}

		t.AppendRows([]table.Row{
			{i, strings.ToUpper(d.Name()), model, fmt.Sprintf("%v / %v", memInfo.Used, memInfo.Total), util},
		})
	}

exit:
	utils.UnmaskPrint(stdout, stderr)
	fmt.Println(t.Render())
}
