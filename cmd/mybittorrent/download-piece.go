package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
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
	fmt.Printf("Sent INTERESTED message\n")

	WaitFor(conn, MsgUnchoke)

	pieceId, err := strconv.Atoi(os.Args[5])
	pieces := getPieces(torrent)

	handleErr(err)

	count := sendPieceRequest(torrent, pieceId, conn)
	fmt.Printf("For Piece : [%d] Sent Requests for Blocks\n", pieceId)

	dataFile := getDataFile(count, pieceId, conn, torrent)

	ok := verifyPiece(dataFile, pieces, pieceId)

	if !ok {
		panic("unequal pieces")
	}

	err = os.WriteFile(os.Args[3], dataFile, os.ModePerm)
	handleErr(err)
	fmt.Printf("Piece %d downloaded to %s\n", pieceId, os.Args[3])
}

func getDataFile(count int, pieceId int, conn net.Conn, torrent Torrent) []byte {
	combinedBlockToPiece := make([]byte, torrent.Info.PiecesLen)
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

func sendPieceRequest(metaInfo Torrent, pieceId int, conn net.Conn) int {
	count := 0
	for byteOffset := 0; byteOffset < int(metaInfo.Info.PiecesLen); byteOffset += BLOCK_SIZE {
		payload := make([]byte, 12)
		length := BLOCK_SIZE

		if byteOffset+BLOCK_SIZE > int(metaInfo.Info.PiecesLen) {
			length = int(metaInfo.Info.PiecesLen) - byteOffset
		}

		binary.BigEndian.PutUint32(payload[0:4], uint32(pieceId))
		binary.BigEndian.PutUint32(payload[4:8], uint32(byteOffset))
		fmt.Println("lenght", length)
		binary.BigEndian.PutUint32(payload[8:], uint32(length))

		_, err := conn.Write(createPeerMessage(MsgRequest, payload))
		handleErr(err)
		count++
	}
	return count
}

func verifyPiece(combinedBlockToPiece []byte, pieces []string, pieceId int) bool {
	sum := sha1.Sum(combinedBlockToPiece)
	return string(sum[:]) == pieces[pieceId]
}
