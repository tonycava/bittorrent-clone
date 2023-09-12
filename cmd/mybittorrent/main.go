package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json" // For converting data to JSON format
	"fmt"           // For printing to the console
	"io/ioutil"
	"os"      // For getting command-line arguments
	"strconv" // For string to integer conversion
	"unicode"
	// For checking digit characters
)

func getStringValue(bencodedString string, offset *int) (string, error) {
	var firstColonIndex int
	for i := *offset; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}
	lengthStr := bencodedString[*offset:firstColonIndex]
	length, err := strconv.Atoi(lengthStr) // Extract the length of the original string
	if err != nil {
		return "", err
	}
	// The reaseon why we need to have the offset here vs in the business logic is because
	// the integer that decides the length of the string needs to be accounted in the offset.
	// its hard to put that extra logic in the switch case.
	// Doing it here i can use lengthStr (which is the length of the string that is the integer that shows the length of the string... i know its alot)
	// Add 1
	// Add the integer length of the string
	*offset += len(lengthStr) + 1 + length                                   // Move the offset forward by the length of the length string + 1 for the colon
	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil // Return a slice of the original string based on the length
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
	integer, err := strconv.Atoi(integerStr) // Extract the length of the original string
	if err != nil {
		return 0, err
	}
	*offset += len(integerStr) // Move the offset forward by the length of the length string
	return integer, nil        // Return a slice of the original string based on the length
}

// decodeBencode takes a bencoded string and decodes it
func decodeBencode(bencodedString string, offset *int) (interface{}, error) {
	// first i check offset is greater than length of bencodedString
	// this indicates that we have reached the end of the string or the offset is out of bounds.
	if *offset >= len(bencodedString) {
		return "", fmt.Errorf("offset %d is out of bounds", *offset)
	}
	switch {
	case unicode.IsDigit(rune(bencodedString[*offset])):
		// if the first character is a digit, then it is a string
		str, err := getStringValue(bencodedString, offset)
		if err != nil {
			return "", err
		}
		return str, nil
	case bencodedString[*offset] == 'i':
		// if the first character is 'i', then it is an integer
		(*offset)++ // move the offset forward by 1
		integer, err := getIntegerValue(bencodedString, offset)
		if err != nil {
			return "", err
		}
		(*offset)++ // move the offset forward by 1
		return integer, nil
	case bencodedString[*offset] == 'l':
		// if the first character is 'l', then it is a list
		(*offset)++ // move the offset forward by 1
		list := []interface{}{}
		for bencodedString[*offset] != 'e' {
			// decode the bencoded string recursively
			decoded, err := decodeBencode(bencodedString, offset)
			if err != nil {
				return "", err
			}
			list = append(list, decoded)
		}
		(*offset)++ // move the offset forward by 1
		return list, nil
	case bencodedString[*offset] == 'd':
		// if the first character is 'd', then it is a dictionary
		(*offset)++ // move the offset forward by 1
		// create a map to store the key-value pairs
		dict := map[string]interface{}{}
		for bencodedString[*offset] != 'e' {
			// decode the key
			key, err := decodeBencode(bencodedString, offset)
			if err != nil {
				return "", err
			}
			// decode the value
			value, err := decodeBencode(bencodedString, offset)
			if err != nil {
				return "", err
			}
			// add the key-value pair to the dictionary
			dict[key.(string)] = value
		}
		(*offset)++ // move the offset forward by 1

		return dict, nil
	default:
		return "", fmt.Errorf("unknown character %c at offset %d", bencodedString[*offset], *offset)
	}
}

func encodeBencode(object interface{}) ([]byte, error) {
	result := make([]byte, 0)
	switch value := object.(type) {
	case int:
		return []byte(fmt.Sprintf("i%de", value)), nil
	case string:
		return []byte(fmt.Sprintf("%d:%s", len(value), value)), nil
	case []interface{}:
		result = append(result, 'l')
		for _, v := range value {
			item, err := encodeBencode(v)
			if err != nil {
				return []byte{}, err
			}
			result = append(result, item...)
		}
		result = append(result, 'e')
		return result, nil
	case map[string]interface{}:
		result = append(result, 'd')
		for k, v := range value {
			result = append(result, []byte(fmt.Sprintf("%d:%s", len(k), k))...)
			temp, err := encodeBencode(v)
			if err != nil {
				return []byte{}, err
			}
			result = append(result, temp...)
		}
		result = append(result, 'e')
		return result, nil
	default:
		fmt.Printf("Unknown type: %v\n", value)
		return result, fmt.Errorf("object is not supported for encoding")
	}
}

func main() {
	// Print debug logs
	// fmt.Println("Logs from your program will appear here!")
	// Get the command ('decode') from command-line arguments
	command := os.Args[1]
	// Execute based on the provided command
	if command == "decode" {
		// Get the bencoded value from command-line arguments
		bencodedValue := os.Args[2]
		offset := 0 // initialize offset to 0
		// Decode the bencoded value
		decoded, err := decodeBencode(bencodedValue, &offset)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Convert the decoded value to JSON and print it
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		// Get the torrent file path from command-line arguments
		torrentFilePath := os.Args[2]
		// Read the torrent file
		torrentFile, err := ioutil.ReadFile(torrentFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		offset := 0
		decoded, err := decodeBencode(string(torrentFile), &offset)
		if err != nil {
			fmt.Println(err)
			return
		}

		infoDict := decoded.(map[string]interface{})
		info := infoDict["info"].(map[string]interface{})

		fmt.Println("Tracker URL:", infoDict["announce"])
		fmt.Println("Length:", info["length"])

		data, err := encodeBencode(info)
		if err != nil {
			fmt.Println(err)
			return
		}

		hasher := sha1.New()
		hasher.Write(data)
		fmt.Println("Info Hash:", hex.EncodeToString(hasher.Sum(nil)))

		fmt.Printf("Piece Hashes:\n")
		hashes := []byte(info["pieces"].(string))
		for i := 0; i < len(hashes); i += 20 {
			fmt.Printf("%s\n", hex.EncodeToString(hashes[i:i+20]))
		}

	} else {
		// Exit the program if the command is unknown
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
