package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net"
	"net/http"
	url2 "net/url"
	"os"
)

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

	} else if command == "handshake" {
		serverAddr := os.Args[3]
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			conn.Close()
			fmt.Println("Connection successful")
			return
		}

		torrent := getTorrentFileInfo(os.Args[2])

		handshake := make([]byte, 0)
		handshake = append(handshake, 19)
		handshake = append(handshake, []byte("BitTorrent protocol")...)
		handshake = append(handshake, make([]byte, 8)...)
		handshake = append(handshake, torrent.getInfoHash()...)
		handshake = append(handshake, []byte("00112233445566778899")...)

		_, err = conn.Write(handshake)
		if err != nil {
			fmt.Println("Error sending data:", err)
			os.Exit(1)
		}

		response := make([]byte, len(handshake))
		_, err = conn.Read(response)

		if err != nil {
			fmt.Println("Error receiving data:", err)
			os.Exit(1)
		}

		fmt.Printf("Peer ID: %s\n", hex.EncodeToString(response[len(response)-20:]))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
