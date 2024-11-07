package utils

import (
	"os"
)

var (
	Stdout = os.Stdout
	Stderr = os.Stderr
)

func MaskPrint() (error, func()) {
	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	if err != nil {
		return err, nil
	}
	os.Stdout = nullFile
	os.Stderr = nullFile

	return nil, func() {
		nullFile.Close()
	}
}
