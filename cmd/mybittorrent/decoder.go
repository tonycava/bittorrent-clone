package main

import (
	"regexp"
	"strconv"
)

func decodeIE(bencodedString string) (interface{}, error) {
	regex := regexp.MustCompile("i(-?[0-9]+)e")
	res := regex.FindStringSubmatch(bencodedString)

	resNbr, err := strconv.Atoi(res[1])
	if err != nil {
		return "", err
	}

	return resNbr, nil
}

func decodeNumberWord(bencodedString string) (interface{}, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
}
