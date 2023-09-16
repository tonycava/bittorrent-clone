package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net/http"
	url2 "net/url"
)

func printInfo(torrent Torrent) {
	fmt.Printf("Tracker URL: %v\n", torrent.Announce)
	fmt.Printf("Length: %v\n", torrent.Info.Length)
	fmt.Printf("Info Hash: %x\n", torrent.getInfoHash())
	fmt.Printf("Piece Length: %v\n", torrent.Info.PiecesLen)
	fmt.Printf("Piece Hashes: \n")
	for i := 0; i < len(torrent.Info.Pieces); i += 20 {
		fmt.Printf("%s\n", hex.EncodeToString([]byte(torrent.Info.Pieces[i:i+20])))
	}
}

func decode(bencodedValue string) string {
	decoded, err := decodeBencode(bencodedValue, new(int))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	jsonOutput, err := json.Marshal(decoded)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(jsonOutput)
}

func getPeers(torrent Torrent) []Peer {
	var infoHash = string(torrent.getInfoHash())

	var url = torrent.Announce + fmt.Sprintf(
		"?info_hash=%s&peer_id=00112233445566778899&port=6881&uploaded=0&downloaded=0&left=%v&compact=1",
		url2.QueryEscape(infoHash), torrent.Info.Length,
	)

	response, err := http.Get(url)
	defer response.Body.Close()

	if err != nil {
		fmt.Println(err)
		return nil
	}

	var trackerResponse TorrentTrackerResponse
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	peers := decodePeers([]byte(trackerResponse.Peers))

	return peers
}
