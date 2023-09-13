package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]
		decoded := getDecode(bencodedValue)
		fmt.Println(decoded)
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
		peerId := makeHandHake(serverAddr)
		fmt.Printf("Peer ID: %s\n", peerId)
	case "download_piece":

		torrent := getTorrentFileInfo(os.Args[4])

		peers := getPeers(torrent)

		serverAddr := peers[0].IP + ":" + peers[0].Port

		tcpAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
		if err != nil {
			fmt.Println("Error resolving address:", err)
			os.Exit(1)
		}

		// Create a TCP connection to the server
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			os.Exit(1)
		}
		defer conn.Close()

		fmt.Println("Connected to server")

		defer conn.Close()

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

		listenForMessages(conn, torrent)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
