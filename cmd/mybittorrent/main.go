package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]
		decoded, _ := decodeBencode(bencodedValue, new(int))
		toDisplay, _ := json.Marshal(decoded)
		fmt.Println(string(toDisplay))
		break
	case "info":
		torrentFilePath := os.Args[2]
		torrent := getTorrentFileInfo(torrentFilePath)

		printInfo(torrent)
	case "peers":
		torrentFilePath := os.Args[2]
		torrent := getTorrentFileInfo(torrentFilePath)
		peers := getPeers(torrent)
		for _, peer := range peers {
			fmt.Printf("%v:%v\n", peer.IP, peer.Port)
		}
	case "handshake":
		serverAddr := os.Args[3]
		torrentFilePath := os.Args[2]
		torrent := getTorrentFileInfo(torrentFilePath)
		connection := getConnections(serverAddr)
		peerId := makeHandHake(connection, torrent)
		fmt.Printf("Peer ID: %s\n", peerId)
	case "download_piece":
		torrentFilePath := os.Args[4]
		torrent := getTorrentFileInfo(torrentFilePath)
		peer := getPeers(torrent)[2]
		serverAddr := fmt.Sprintf("%s:%s", peer.IP, peer.Port)

		conn := getConnections(serverAddr)
		defer conn.Close()

		fmt.Println("Connected to server")

		makeHandHake(conn, torrent)
		listenForMessages(conn, torrent)
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
