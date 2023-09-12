package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json" // For converting data to JSON format
	"fmt"           // For printing to the console
	"github.com/jackpal/bencode-go"
	"io/ioutil"
	"os"      // For getting command-line arguments
	"strconv" // For string to integer conversion
	"unicode"
	// For checking digit characters
)

type Torrent struct {
	Announce string `json:"announce" bencode:"announce"`
	Info     Info   `json:"info" bencode:"info"`
}
type Info struct {
	Length    int64  `json:"length" bencode:"length"`
	Name      string `json:"name" bencode:"name"`
	PiecesLen int64  `json:"piece length" bencode:"piece length"`
	Pieces    string `json:"pieces" bencode:"pieces"`
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
		content, err := ioutil.ReadFile(torrentFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		var torrent Torrent
		if err := bencode.Unmarshal(bytes.NewReader(content), &torrent); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		var buffer_ bytes.Buffer

		if err := bencode.Marshal(&buffer_, torrent.Info); err != nil {
			fmt.Println("Error marshalling BEncode:", err)
			return
		}

		hash := sha1.Sum(buffer_.Bytes())
		fmt.Printf("Tracker URL: %v\n", torrent.Announce)
		fmt.Printf("Length: %v\n", torrent.Info.Length)
		fmt.Printf("Info Hash: %x\n", hash)
		fmt.Printf("Piece Length: %v\n", torrent.Info.PiecesLen)
		fmt.Printf("Piece Hashes: \n")
		i := 0
		for ; i < len(torrent.Info.Pieces)/20; i++ {
			piece := torrent.Info.Pieces[i*20 : (i*20)+20]
			fmt.Printf("%x\n", piece)
		}
		if len(torrent.Info.Pieces) > i*20 {
			fmt.Println(torrent.Info.Pieces[i*20+1])
		}

	} else {
		// Exit the program if the command is unknown
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
