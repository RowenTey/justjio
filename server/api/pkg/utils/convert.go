package utils

import "strconv"

func UIntToString(num uint) string {
	return strconv.FormatUint(uint64(num), 10)
}

func StringToUint(str string) (uint, error) {
	num, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}
