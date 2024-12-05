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

// StringInSlice returns true if s is in list
func StringInSlice(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}

	return false
}
