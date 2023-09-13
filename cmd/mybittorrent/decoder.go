package main

import (
	"fmt"
	"strconv"
	"unicode"
)

func decodeBencode(bencodedString string, offset *int) (interface{}, error) {
	if *offset >= len(bencodedString) {
		return "", fmt.Errorf("offset %d is out of bounds", *offset)
	}

	switch {
	case unicode.IsDigit(rune(bencodedString[*offset])):
		str, err := getStringValue(bencodedString, offset)
		if err != nil {
			return "", err
		}
		return str, nil
	case bencodedString[*offset] == 'i':
		*offset += 1
		integer, err := getIntegerValue(bencodedString, offset)
		if err != nil {
			return "", err
		}
		*offset += 1
		return integer, nil
	case bencodedString[*offset] == 'l':
		*offset += 1
		list := []interface{}{}
		for bencodedString[*offset] != 'e' {
			decoded, err := decodeBencode(bencodedString, offset)
			if err != nil {
				return "", err
			}
			list = append(list, decoded)
		}
		*offset += 1
		return list, nil
	case bencodedString[*offset] == 'd':
		*offset += 1
		dict := map[string]interface{}{}
		for bencodedString[*offset] != 'e' {
			key, err := decodeBencode(bencodedString, offset)
			if err != nil {
				return "", err
			}
			value, err := decodeBencode(bencodedString, offset)
			if err != nil {
				return "", err
			}
			dict[key.(string)] = value
		}
		*offset += 1
		return dict, nil
	default:
		return "", fmt.Errorf("unknown character %c at offset %d", bencodedString[*offset], *offset)
	}
}

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
	var firstEIdx int
	for i := *offset; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			firstEIdx = i
			break
		}
	}
	integerStr := bencodedString[*offset:firstEIdx]
	integer, err := strconv.Atoi(integerStr)
	if err != nil {
		return 0, err
	}
	*offset += len(integerStr)
	return integer, nil
}
