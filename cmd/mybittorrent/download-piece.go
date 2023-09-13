package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func listenForMessages(conn net.Conn, torrent Torrent) {
	WaitFor(conn, MsgBitfield)

	_, err := conn.Write(createPeerMessage(MsgInterested, []byte{}))
	if err != nil {
		fmt.Println("Error sending request message:", err)
		return
	}

	WaitFor(conn, MsgUnchoke)

	pieceId, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatal(err)
	}
	count := 0

	for byteOffset := 0; byteOffset < int(torrent.Info.PiecesLen); byteOffset = byteOffset + BLOCK_SIZE {
		payload := make([]byte, 12)
		binary.BigEndian.PutUint32(payload[0:4], uint32(pieceId))
		binary.BigEndian.PutUint32(payload[4:8], uint32(byteOffset))
		binary.BigEndian.PutUint32(payload[8:], BLOCK_SIZE)
		_, err = conn.Write(createPeerMessage(MsgRequest, payload))
		if err != nil {
			fmt.Println("Error sending request message:", err)
			return
		}
		count++
	}

	dataFile := getDataFile(count, pieceId, conn, torrent)
	writeFile(os.Args[3], dataFile)
	fmt.Printf("Piece %d downloaded to %s.\n", pieceId, os.Args[3])
}

func getDataFile(count int, pieceId int, conn net.Conn, torrent Torrent) []byte {
	combinedBlockToPiece := make([]byte, torrent.Info.Length)
	for i := 0; i < count; i++ {
		data := WaitFor(conn, MsgPiece)
		index := binary.BigEndian.Uint32(data[0:4])
		if index != uint32(pieceId) {
			panic(fmt.Sprintf("something went wrong [expected: %d -- actual: %d]", pieceId, index))
		}
		begin := binary.BigEndian.Uint32(data[4:8])
		block := data[8:]
		copy(combinedBlockToPiece[begin:], block)
	}
	return combinedBlockToPiece
}
