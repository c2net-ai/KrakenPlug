package utils

import (
	"regexp"
	"strconv"
)

func JoinSliceInt(idxs []int) string {
	idxsStr := ""
	for i, idx := range idxs {
		if i != 0 {
			idxsStr += ","
		}
		idxsStr += strconv.Itoa(idx)
	}

	return idxsStr
}

func ReplaceAllBlank(input string) string {
	re := regexp.MustCompile(`\s+`)

	result := re.ReplaceAllString(input, "-")

	return result
}
