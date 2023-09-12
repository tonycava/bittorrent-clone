package main

import (
	"encoding/json"
	"reflect"

	// Uncomment this line to pass the first stage
	// "encoding/json"
	"fmt"
	"os"
	"unicode"
)

func IsSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeNumberWord(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeIE(bencodedString)
	} else if bencodedString[0] == 'l' {
		if bencodedString == "le" {
			return []interface{}{}, nil
		}
		start := bencodedString[1 : len(bencodedString)-1]
		var finalList []interface{}
		var toDecode string
		var isAtFirstCase = false
		var isAtSecondCase = false

		for start[0] == 'l' && start[len(start)-1] == 'e' {
			finalList = append(finalList, []interface{}{})
			start = start[1 : len(start)-1]
		}

		for i := 0; i < len(start); i++ {
			toDecode += string(start[i])

			if isAtFirstCase && start[i] == 'i' {
				var idx = indexOfSemi(toDecode)
				var word = toDecode[idx+1 : len(toDecode)-1]
				res, _ := decodeNumberWord(fmt.Sprintf("%v:%v", len(word), word))
				if len(finalList) >= 1 && IsSlice(finalList[0]) {
					finalList[0] = append(finalList[0].([]interface{}), res)
				} else {
					finalList = append(finalList, res)
				}

				toDecode = ""
				isAtSecondCase = false
				isAtFirstCase = false
			}

			if isAtSecondCase && start[i] == 'e' {
				if toDecode[0] == 'i' {
					res, _ := decodeIE(toDecode)

					if len(finalList) >= 1 && IsSlice(finalList[0]) {
						finalList[0] = append(finalList[0].([]interface{}), res)
					} else {
						finalList = append(finalList, res)
					}

					toDecode = ""

					isAtSecondCase = false
					isAtFirstCase = false
					continue
				}

				res, _ := decodeIE("i" + toDecode)

				if len(finalList) >= 1 && IsSlice(finalList[0]) {
					finalList[0] = append(finalList[0].([]interface{}), res)
				} else {
					finalList = append(finalList, res)
				}

				toDecode = ""
				isAtSecondCase = false
				isAtFirstCase = false
				continue
			}

			if start[i] == ':' {
				isAtFirstCase = true
			}

			if start[i] == 'i' {
				isAtSecondCase = true
			}

		}

		if toDecode != "" {
			res, _ := decodeNumberWord(toDecode)

			if len(finalList) >= 1 && IsSlice(finalList[0]) {
				finalList[0] = append(finalList[0].([]interface{}), res)
			} else {
				finalList = append(finalList, res)
			}
		}
		return finalList, nil
	} else if bencodedString[0] == 'd' {
		if bencodedString == "de" {
			return map[string]interface{}{}, nil
		}
		start := bencodedString[1 : len(bencodedString)-1]
		var toDecode string
		var finalMap = make(map[string]interface{})
		var isAtFirstCase = false
		var isAtSecondCase = false
		var firstWord string

		for i := 0; i < len(start); i++ {
			toDecode += string(start[i])

			if isAtFirstCase && (start[i] == ':' || unicode.IsDigit(rune(start[i]))) {
				if toDecode != "" {
					var idx = indexOfSemi(toDecode)
					var word = toDecode[idx+1 : len(toDecode)-1]

					if firstWord != "" {
						finalMap[firstWord] = word
						firstWord = ""
					} else {
						firstWord = word
					}
					toDecode = ""
				}

				isAtSecondCase = false
				isAtFirstCase = false
			}

			if isAtSecondCase && start[i] == 'i' {
				if toDecode != "" {
					var idx = indexOfSemi(toDecode)
					var word = toDecode[idx+1 : len(toDecode)-1]

					if firstWord != "" {
						finalMap[firstWord] = word
						firstWord = ""
					} else {
						firstWord = word
					}

					toDecode = ""
				}

				isAtSecondCase = false
				isAtFirstCase = false
				toDecode = ""
			}

			if start[i] == ':' {
				isAtFirstCase = true
			}

			if start[i] == 'e' {
				isAtSecondCase = true
			}
		}
		if toDecode[0] != 'i' {
			res, _ := decodeIE("i" + toDecode)

			finalMap[firstWord] = res
		} else {
			res, _ := decodeIE(toDecode)

			finalMap[firstWord] = res
		}
		return finalMap, nil

	}
	return "", fmt.Errorf("Only strings are supported at the moment")

}

func indexOfSemi(word string) int {
	for i := 0; i < len(word); i++ {
		if word[i] == ':' {
			return i
		}
	}
	return -1
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
