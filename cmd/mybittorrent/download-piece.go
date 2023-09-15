package main

import (
	"crypto/sha1"
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
	fmt.Printf("Sent INTERESTED message\n")

	WaitFor(conn, MsgUnchoke)

	pieceId, err := strconv.Atoi(os.Args[5])
	pieces := getPieces(torrent)

	handleErr(err)

	sendPieceRequest(torrent, pieceId, conn)
	fmt.Printf("For Piece : [%d] Sent Requests for Blocks\n", pieceId)

	pieceLength := torrent.Info.PiecesLen
	if pieceId == len(torrent.Info.Pieces)-1 {
		pieceLength = torrent.Info.Length - (int64(pieceId) * torrent.Info.PiecesLen)
	}
	lastBlockSize := pieceLength % int64(BLOCK_SIZE)
	numBlocks := (pieceLength - lastBlockSize) / int64(BLOCK_SIZE)
	log.Printf("there are %d blocks in piece %d\n", numBlocks, pieceId)
	if lastBlockSize > 0 {
		log.Printf("piece %d has an unaligned block of size %d\n", pieceId, lastBlockSize)
		numBlocks++
	} else {
		log.Printf("piece %d has size of %d and is aligned with blocksize of %d\n", pieceId, torrent.Info.PiecesLen, BLOCK_SIZE)
	}

	dataFile := getDataFile(int(numBlocks), pieceId, conn, torrent)

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
	for byteOffset := 0; byteOffset < int(metaInfo.Info.PiecesLen); byteOffset = byteOffset + BLOCK_SIZE {
		payload := make([]byte, 12)
		binary.BigEndian.PutUint32(payload[0:4], uint32(pieceId))
		binary.BigEndian.PutUint32(payload[4:8], uint32(byteOffset))
		binary.BigEndian.PutUint32(payload[8:], BLOCK_SIZE)

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
