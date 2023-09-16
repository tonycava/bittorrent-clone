package main

import "fmt"

const BLOCK_SIZE = 16 * 1024

const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8
)

func getPieces(metaInfo Torrent) []string {
	pieces := make([]string, len(metaInfo.Info.Pieces)/20)
	for i := 0; i < len(metaInfo.Info.Pieces)/20; i++ {
		piece := metaInfo.Info.Pieces[i*20 : (i*20)+20]
		pieces[i] = piece
	}
	return pieces
}

func handleErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
}
