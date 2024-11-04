package utils

import (
	"os"
)

func MaskPrint() (error, *os.File, *os.File, func()) {
	stdout := os.Stdout
	stderr := os.Stderr
	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	if err != nil {
		return err, nil, nil, nil
	}
	os.Stdout = nullFile
	os.Stderr = nullFile

	return nil, stdout, stderr, func() {
		nullFile.Close()
	}
}

func UnmaskPrint(stdout *os.File, stderr *os.File) {
	os.Stdout = stdout
	os.Stderr = stderr
}
