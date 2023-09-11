package main

import (
	"encoding/json"
	// Uncomment this line to pass the first stage
	// "encoding/json"
	"fmt"
	"os"
	"unicode"
)

func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeNumberWord(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeIE(bencodedString)
	} else if bencodedString[0] == 'l' {
		start := bencodedString[1 : len(bencodedString)-1]
		var finalList []interface{}
		var toDecode string
		var isAtFirstCase = false
		var isAtSecondCase = false

		for i := 0; i < len(start); i++ {
			toDecode += string(start[i])

			if isAtFirstCase && start[i] == 'i' {
				var idx = indexOfSemi(toDecode)
				var word = toDecode[idx+1 : len(toDecode)-1]
				res, _ := decodeNumberWord(fmt.Sprintf("%v:%v", len(word), word))
				finalList = append(finalList, res)
				toDecode = ""
				isAtSecondCase = false
				isAtFirstCase = false
			}

			if isAtSecondCase && start[i] == 'e' {
				if toDecode[0] == 'i' {
					res, _ := decodeIE(toDecode)
					finalList = append(finalList, res)
					toDecode = ""

					isAtSecondCase = false
					isAtFirstCase = false
					continue
				}

				res, _ := decodeIE("i" + toDecode)
				finalList = append(finalList, res)

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
		return finalList, nil
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
