package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
)

type Torrent struct {
	Announce string `json:"announce" bencode:"announce"`
	Info     Info   `json:"info" bencode:"info"`
}

//type Info struct {
//	Length    int64  `json:"length" bencode:"length"`
//	Name      string `json:"name" bencode:"name"`
//	PiecesLen int64  `json:"piece length" bencode:"piece length"`
//	Pieces    string `json:"pieces" bencode:"pieces"`
//}

//type Peer struct {
//	IP   string `json:"ip"`
//	Port string `json:"port"`
//}

type TorrentTrackerResponse struct {
	Interval int    `json:"interval" bencode:"interval"`
	Peers    string `json:"peers" bencode:"peers"`
}

func (t Torrent) getInfoHash() []byte {
	hasher := sha1.New()
	if err := bencode.Marshal(hasher, t.Info); err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return hasher.Sum(nil)
}
