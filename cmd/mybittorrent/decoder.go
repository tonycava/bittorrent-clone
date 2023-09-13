package main

import "strconv"

func getStringValue(bencodedString string, offset *int) (string, error) {
	var firstColonIndex int
	for i := *offset; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}
	lengthStr := bencodedString[*offset:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}
	*offset += len(lengthStr) + 1 + length
	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
}
func getIntegerValue(bencodedString string, offset *int) (int, error) {
	var firstColonIndex int
	for i := *offset; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			firstColonIndex = i
			break
		}
	}
	integerStr := bencodedString[*offset:firstColonIndex]
	integer, err := strconv.Atoi(integerStr)
	if err != nil {
		return 0, err
	}
	*offset += len(integerStr)
	return integer, nil
}
