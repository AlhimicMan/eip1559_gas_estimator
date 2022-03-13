package estimator

import (
	"strconv"
)

func hexToInt(hexVal string) (int64, error) {
	hexValStr := hexVal[2:]
	value, err := strconv.ParseInt(hexValStr, 16, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}
