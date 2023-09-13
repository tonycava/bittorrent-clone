package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io/ioutil"
)

func getTorrentFileInfo(path string) Torrent {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return Torrent{}
	}

	var torrent Torrent
	if err := bencode.Unmarshal(bytes.NewReader(content), &torrent); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return Torrent{}
	}

	return torrent
}

type Peer struct {
	IP   string
	Port string
}

func getPeers(peers []byte) []Peer {
	var result []Peer
	for i := 0; i < len(peers); i += 6 {
		result = append(result, Peer{
			IP:   fmt.Sprintf("%d.%d.%d.%d", peers[i], peers[i+1], peers[i+2], peers[i+3]),
			Port: fmt.Sprintf("%d", binary.BigEndian.Uint16(peers[i+4:i+6])),
		})
	}

	return result
}
