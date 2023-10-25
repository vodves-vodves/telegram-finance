package bot

import (
	"strconv"
)

func isNumeric(s string) (int, bool) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, true
	}
	return 0, false
}
