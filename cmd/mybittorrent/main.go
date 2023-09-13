package main

import (
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net/http"
	url2 "net/url"
	"os"
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

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue, new(int))
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(jsonOutput))

	} else if command == "info" {
		torrentFilePath := os.Args[2]
		torrent := getTorrentFileInfo(torrentFilePath)

		fmt.Printf("Tracker URL: %v\n", torrent.Announce)
		fmt.Printf("Length: %v\n", torrent.Info.Length)
		fmt.Printf("Info Hash: %x\n", torrent.getInfoHash())
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

	} else if command == "peers" {
		torrentFilePath := os.Args[2]
		torrent := getTorrentFileInfo(torrentFilePath)

		var infoHash = string(torrent.getInfoHash())

		var url = torrent.Announce + fmt.Sprintf(
			"?info_hash=%s&peer_id=00112233445566778899&port=6881&uploaded=0&downloaded=0&left=%v&compact=1",
			url2.QueryEscape(infoHash), torrent.Info.Length,
		)

		response, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
			return
		}

		var trackerResponse TorrentTrackerResponse
		err = bencode.Unmarshal(response.Body, &trackerResponse)
		if err != nil {
			fmt.Println(err)
			return
		}

		peers := getPeers([]byte(trackerResponse.Peers))
		for i := 0; i < len(peers); i++ {
			fmt.Println(peers[i].IP + ":" + peers[i].Port)
		}

		defer response.Body.Close()

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
