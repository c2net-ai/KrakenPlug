package utils

import "strconv"

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
